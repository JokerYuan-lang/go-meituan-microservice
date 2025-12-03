package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductRepo 商品数据访问接口
type ProductRepo interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, productID, merchantID int64) error
	ListProductsByMerchantID(ctx context.Context, merchantID int64, page, pageSize int32) ([]*model.Product, int64, error)
	GetProductByID(ctx context.Context, productID int64) (*model.Product, error)
	DeductStock(ctx context.Context, productID int64, num int32) error  // 扣减库存（悲观锁）
	RestoreStock(ctx context.Context, productID int64, num int32) error // 恢复库存
}

// productRepo 实现
type productRepo struct{}

// NewProductRepo 创建实例
func NewProductRepo() ProductRepo {
	return &productRepo{}
}

func (p *productRepo) CreateProduct(ctx context.Context, product *model.Product) error {
	tx := db.Mysql.WithContext(ctx).Create(product)
	if tx.Error != nil {
		zap.L().Error("创建商品失败", zap.Any("product", product), zap.Error(tx.Error))
		return utils.NewDBError("创建商品失败：" + tx.Error.Error())
	}
	return nil
}

func (p *productRepo) UpdateProduct(ctx context.Context, product *model.Product) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Product{}).Where("product_id = ? AND merchant_id = ?", product.ProductID, product.MerchantID).Updates(map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"stock":       product.Stock,
		"image_url":   product.ImageURL,
		"is_sold_out": product.IsSoldOut,
	})
	if tx.Error != nil {
		zap.L().Error("更新商品失败", zap.Any("product", product), zap.Error(tx.Error))
		return utils.NewDBError("更新商品失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("商品不存在或无权限更新")
	}
	return nil
}

func (p *productRepo) DeleteProduct(ctx context.Context, productID, merchantID int64) error {
	tx := db.Mysql.WithContext(ctx).Where("product_id = ? AND merchant_id = ?", productID, merchantID).Delete(&model.Product{})
	if tx.Error != nil {
		zap.L().Error("删除商品失败", zap.Error(tx.Error))
		return utils.NewDBError("删除商品失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("商品不存在或无权限删除")
	}
	return nil
}

func (p *productRepo) ListProductsByMerchantID(ctx context.Context, merchantID int64, page, pageSize int32) ([]*model.Product, int64, error) {
	var (
		total    int64
		products []*model.Product
	)
	//先求数量
	err := db.Mysql.WithContext(ctx).Model(&model.Product{}).Where("merchant_id = ?", merchantID).Count(&total).Error
	if err != nil {
		zap.L().Error("统计商品总数失败", zap.Int64("merchant_id", merchantID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询商品失败：" + err.Error())
	}
	offset := int((page - 1) * pageSize)
	tx := db.Mysql.WithContext(ctx).Model(&model.Product{}).Where("merchant_id = ?", merchantID).Offset(offset).Limit(int(pageSize)).Order("updated_at desc").Find(&products)
	if tx.Error != nil {
		zap.L().Error("查询商品列表失败", zap.Int64("merchant_id", merchantID), zap.Error(tx.Error))
		return nil, 0, utils.NewDBError("查询商品失败：" + tx.Error.Error())
	}
	return products, total, nil
}

func (p *productRepo) GetProductByID(ctx context.Context, productID int64) (*model.Product, error) {
	product := &model.Product{}
	tx := db.Mysql.WithContext(ctx).Where("product_id = ?", productID).First(&product)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError("商品不存在")
		}
		zap.L().Error("通过商品ID查询商品失败", zap.Int64("product_id", productID), zap.Error(tx.Error))
		return nil, utils.NewDBError("通过商品ID查询商品失败：" + tx.Error.Error())
	}
	return product, nil
}

// DeductStock 扣减库存
func (p *productRepo) DeductStock(ctx context.Context, productID int64, num int32) error {
	tx := db.Mysql.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var product model.Product
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id = ?", productID).First(&product).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewBizError("商品不存在")
		}
		return utils.NewDBError("扣减库存失败" + err.Error())
	}
	//校验库存
	if product.Stock < num {
		tx.Rollback()
		return utils.NewBizError("库存不足")
	}
	product.Stock -= num
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		zap.L().Error("扣减库存失败", zap.Int64("product_id", productID), zap.Int32("num", num), zap.Error(err))
		return utils.NewDBError("扣减库存失败：" + err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.NewDBError("扣减库存失败：" + err.Error())
	}
	return nil
}

func (p *productRepo) RestoreStock(ctx context.Context, productID int64, num int32) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Product{}).Where("product_id = ?", productID).Update("stock", gorm.Expr("stock + ?", num))
	if tx.Error != nil {
		zap.L().Error("恢复库存失败", zap.Int64("product_id", productID), zap.Int32("num", num), zap.Error(tx.Error))
		return utils.NewDBError("恢复库存失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("商品不存在")
	}
	return nil
}
