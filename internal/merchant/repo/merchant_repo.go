package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MerchantRepo 商家数据访问接口
type MerchantRepo interface {
	CreateMerchant(ctx context.Context, merchant *model.Merchant) error
	GetMerchantByPhone(ctx context.Context, phone string) (*model.Merchant, error)
	GetMerchantByID(ctx context.Context, merchantID int64) (*model.Merchant, error)
	UpdateMerchant(ctx context.Context, merchant *model.Merchant) error
	UpdateOrderCount(ctx context.Context, merchantID int64, num int32) error // 更新订单数
}

// merchantRepo 实现
type merchantRepo struct{}

// NewMerchantRepo 创建实例
func NewMerchantRepo() MerchantRepo {
	return &merchantRepo{}
}

// CreateMerchant 商家入驻（创建商家）
func (r *merchantRepo) CreateMerchant(ctx context.Context, merchant *model.Merchant) error {
	tx := db.Mysql.WithContext(ctx).Create(merchant)
	if tx.Error != nil {
		zap.L().Error("创建商家失败", zap.Any("merchant", merchant), zap.Error(tx.Error))
		return utils.NewDBError("商家入驻失败：" + tx.Error.Error())
	}
	return nil
}

// GetMerchantByPhone 根据手机号查询商家（登录用）
func (r *merchantRepo) GetMerchantByPhone(ctx context.Context, phone string) (*model.Merchant, error) {
	var merchant model.Merchant
	tx := db.Mysql.WithContext(ctx).Where("phone = ?", phone).First(&merchant)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			zap.L().Warn("商家不存在（手机号）", zap.String("phone", phone))
			return nil, nil // 不存在返回nil
		}
		zap.L().Error("查询商家失败（手机号）", zap.String("phone", phone), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询商家失败：" + tx.Error.Error())
	}
	return &merchant, nil
}

// GetMerchantByID 根据ID查询商家信息
func (r *merchantRepo) GetMerchantByID(ctx context.Context, merchantID int64) (*model.Merchant, error) {
	var merchant model.Merchant
	tx := db.Mysql.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&merchant)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			zap.L().Warn("商家不存在（ID）", zap.Int64("merchant_id", merchantID))
			return nil, utils.NewBizError("商家不存在")
		}
		zap.L().Error("查询商家失败（ID）", zap.Int64("merchant_id", merchantID), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询商家失败：" + tx.Error.Error())
	}
	return &merchant, nil
}

// UpdateMerchant 更新商家信息
func (r *merchantRepo) UpdateMerchant(ctx context.Context, merchant *model.Merchant) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Merchant{}).
		Where("merchant_id = ?", merchant.MerchantID).
		Updates(map[string]interface{}{
			"name":           merchant.Name,
			"phone":          merchant.Phone,
			"address":        merchant.Address,
			"logo":           merchant.Logo,
			"business_hours": merchant.BusinessHours,
			"is_open":        merchant.IsOpen,
		})
	if tx.Error != nil {
		zap.L().Error("更新商家信息失败", zap.Any("merchant", merchant), zap.Error(tx.Error))
		return utils.NewDBError("更新商家信息失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("商家不存在")
	}
	return nil
}

// UpdateOrderCount 更新商家订单数（接单时+1）
func (r *merchantRepo) UpdateOrderCount(ctx context.Context, merchantID int64, num int32) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Merchant{}).
		Where("merchant_id = ?", merchantID).
		Update("order_count", gorm.Expr("order_count + ?", num))
	if tx.Error != nil {
		zap.L().Error("更新商家订单数失败", zap.Int64("merchant_id", merchantID), zap.Int32("num", num), zap.Error(tx.Error))
		return utils.NewDBError("更新订单数失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("商家不存在")
	}
	return nil
}
