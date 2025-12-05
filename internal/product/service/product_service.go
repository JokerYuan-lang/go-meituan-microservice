package service

import (
	"context"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type CreateProductParam struct {
	MerchantID  int64   `validate:"required,gt=0"`
	Name        string  `validate:"required,min=2,max=64"`
	Description string  `validate:"max=512"`
	Price       float64 `validate:"required,gt=0"`
	Stock       int32   `validate:"required,gte=0"`
	ImageURL    string  `validate:"required,url"`
}

type UpdateProductParam struct {
	ProductID   int64   `validate:"required,gt=0"`
	MerchantID  int64   `validate:"required,gt=0"`
	Name        string  `validate:"required,min=2,max=64"`
	Description string  `validate:"max=512"`
	Price       float64 `validate:"required,gt=0"`
	Stock       int32   `validate:"required,gte=0"`
	ImageURL    string  `validate:"required,url"`
	IsSoldOut   bool    `validate:"required"`
}

type DeleteProductParam struct {
	ProductID  int64 `validate:"required,gt=0"`
	MerchantID int64 `validate:"required,gt=0"`
}

type ListProductsParam struct {
	MerchantID int64 `validate:"required,gt=0"`
	Page       int32 `validate:"required,gte=1"`
	PageSize   int32 `validate:"required,gte=10,lte=100"`
}

type DeductStockParam struct {
	ProductID int64 `validate:"required,gt=0"`
	Num       int32 `validate:"required,gt=0"`
}

type RestoreStockParam struct {
	ProductID int64 `validate:"required,gt=0"`
	Num       int32 `validate:"required,gt=0"`
}

// 响应结构体（领域层）
type ProductResult struct {
	ProductID   int64   `json:"product_id"`
	MerchantID  int64   `json:"merchant_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
	IsSoldOut   bool    `json:"is_sold_out"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListProductsResult struct {
	Products []ProductResult `json:"products"`
	Total    int64           `json:"total"`
	Page     int32           `json:"page"`
	PageSize int32           `json:"page_size"`
}

// ProductService 商品业务逻辑接口
type ProductService interface {
	CreateProduct(ctx context.Context, param CreateProductParam) (int64, error)
	UpdateProduct(ctx context.Context, param UpdateProductParam) error
	DeleteProduct(ctx context.Context, param DeleteProductParam) error
	ListProductsByMerchantID(ctx context.Context, param ListProductsParam) (ListProductsResult, error)
	GetProductByID(ctx context.Context, productID int64) (ProductResult, error)
	DeductStock(ctx context.Context, param DeductStockParam) error
	RestoreStock(ctx context.Context, param RestoreStockParam) error
}

// productService 实现
type productService struct {
	productRepo repo.ProductRepo
	validate    *validator.Validate
}

// NewProductService 创建实例
func NewProductService(productRepo repo.ProductRepo) ProductService {
	return &productService{
		productRepo: productRepo,
		validate:    validator.New(),
	}
}

func (s *productService) CreateProduct(ctx context.Context, param CreateProductParam) (int64, error) {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("创建商品参数校验失败", zap.Error(err))
		return 0, utils.NewParamError("创建商品参数校验失败" + err.Error())
	}
	product := &model.Product{
		MerchantID:  param.MerchantID,
		Name:        param.Name,
		Description: param.Description,
		Price:       param.Price,
		Stock:       param.Stock,
		ImageURL:    param.ImageURL,
		IsSoldOut:   param.Stock <= 0,
	}
	err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return 0, err
	}
	zap.L().Info("创建商品成功", zap.Any("product", product))
	return product.ProductID, nil
}

func (s *productService) UpdateProduct(ctx context.Context, param UpdateProductParam) error {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("更新商品参数校验失败", zap.Error(err))
		return utils.NewParamError("更新商品参数校验失败" + err.Error())
	}
	product := &model.Product{
		MerchantID:  param.MerchantID,
		Name:        param.Name,
		Description: param.Description,
		Price:       param.Price,
		Stock:       param.Stock,
		ImageURL:    param.ImageURL,
		IsSoldOut:   param.IsSoldOut,
	}
	err := s.productRepo.UpdateProduct(ctx, product)
	if err != nil {
		return err
	}
	zap.L().Info("更新商品成功", zap.Any("product", product))
	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, param DeleteProductParam) error {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("删除商品参数校验失败", zap.Error(err))
		return utils.NewParamError("删除商品参数校验失败" + err.Error())
	}
	if err := s.productRepo.DeleteProduct(ctx, param.ProductID, param.MerchantID); err != nil {
		return err
	}
	zap.L().Info("删除商品成功", zap.Int64("product", param.ProductID), zap.Int64("merchant", param.MerchantID))
	return nil
}

func (s *productService) ListProductsByMerchantID(ctx context.Context, param ListProductsParam) (ListProductsResult, error) {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询商品列表参数校验错误", zap.Error(err))
		return ListProductsResult{}, utils.NewParamError("查询商品列表参数校验错误" + err.Error())
	}
	products, total, err := s.productRepo.ListProductsByMerchantID(ctx, param.MerchantID, param.Page, param.PageSize)
	if err != nil {
		return ListProductsResult{}, err
	}
	productsResult := make([]ProductResult, len(products))
	for _, product := range products {
		productsResult = append(productsResult, ProductResult{
			ProductID:   product.ProductID,
			MerchantID:  product.MerchantID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			ImageURL:    product.ImageURL,
			IsSoldOut:   product.IsSoldOut,
			CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return ListProductsResult{
		Products: productsResult,
		Total:    total,
		Page:     param.Page,
		PageSize: param.PageSize,
	}, nil
}

func (s *productService) GetProductByID(ctx context.Context, productID int64) (ProductResult, error) {
	if productID <= 0 {
		zap.L().Warn("商品ID不能为空")
		return ProductResult{}, utils.NewParamError("商品ID为空")
	}
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return ProductResult{}, err
	}

	return ProductResult{
		ProductID:   product.ProductID,
		MerchantID:  product.MerchantID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		IsSoldOut:   product.IsSoldOut,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *productService) DeductStock(ctx context.Context, param DeductStockParam) error {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("删减库存参数校验失败", zap.Error(err))
		return utils.NewParamError("删减库存参数校验失败" + err.Error())
	}
	err := s.productRepo.DeductStock(ctx, param.ProductID, param.Num)
	if err != nil {
		return err
	}
	return nil
}

func (s *productService) RestoreStock(ctx context.Context, param RestoreStockParam) error {
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("增加库存参数校验失败", zap.Error(err))
		return utils.NewParamError("增加库存参数校验失败" + err.Error())
	}
	err := s.productRepo.RestoreStock(ctx, param.ProductID, param.Num)
	if err != nil {
		return err
	}
	return nil
}
