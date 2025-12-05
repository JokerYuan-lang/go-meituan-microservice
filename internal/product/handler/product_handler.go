package handler

import (
	"context"
	"errors"

	productProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/product/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/service"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
)

// ProductHandler 商品gRPC接口实现
type ProductHandler struct {
	productProto.UnimplementedProductServiceServer
	productService service.ProductService
}

// NewProductHandler 创建实例
func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (p *ProductHandler) CreateProduct(ctx context.Context, req *productProto.CreateProductRequest) (*productProto.CreateProductResponse, error) {
	param := service.CreateProductParam{
		MerchantID:  req.MerchantId,
		Name:        req.Name,
		Description: req.Description,
		Price:       float64(req.Price),
		Stock:       req.Stock,
		ImageURL:    req.ImageUrl,
	}

	productID, err := p.productService.CreateProduct(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("创建商品未知错误", zap.Error(err))
			return &productProto.CreateProductResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.CreateProductResponse{
			Code:      int32(appError.Code),
			Msg:       appError.Message,
			ProductId: 0,
		}, nil
	}
	return &productProto.CreateProductResponse{
		Code:      utils.ErrCodeSuccess,
		Msg:       "创建商品成功",
		ProductId: productID,
	}, nil
}

func (p *ProductHandler) UpdateProduct(ctx context.Context, req *productProto.UpdateProductRequest) (*productProto.CommonResponse, error) {
	param := service.UpdateProductParam{
		ProductID:   req.ProductId,
		MerchantID:  req.MerchantId,
		Name:        req.Name,
		Description: req.Description,
		Price:       float64(req.Price),
		Stock:       req.Stock,
		ImageURL:    req.ImageUrl,
		IsSoldOut:   req.IsSoldOut,
	}
	err := p.productService.UpdateProduct(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("更新商品未知错误", zap.Error(err))
			return &productProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.CommonResponse{
			Code: int32(appError.Code),
			Msg:  appError.Message,
		}, nil
	}
	return &productProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新商品成功",
	}, nil
}

func (p *ProductHandler) DeleteProduct(ctx context.Context, req *productProto.DeleteProductRequest) (*productProto.CommonResponse, error) {
	param := service.DeleteProductParam{
		MerchantID: req.MerchantId,
		ProductID:  req.ProductId,
	}
	err := p.productService.DeleteProduct(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("删除商品未知错误", zap.Error(err))
			return &productProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.CommonResponse{
			Code: int32(appError.Code),
			Msg:  appError.Message,
		}, nil
	}
	return &productProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "删除商品成功",
	}, nil
}

func (p *ProductHandler) ListProductByMerchantID(ctx context.Context, req *productProto.ListProductsRequest) (*productProto.ListProductsResponse, error) {
	param := service.ListProductsParam{
		MerchantID: req.MerchantId,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	ListProductsResult, err := p.productService.ListProductsByMerchantID(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("查询商品列表未知错误", zap.Error(err))
			return &productProto.ListProductsResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.ListProductsResponse{
			Code: int32(appError.Code),
			Msg:  appError.Message,
		}, nil
	}
	var products []*productProto.Product
	for _, productResult := range ListProductsResult.Products {
		products = append(products, &productProto.Product{
			ProductId:   productResult.ProductID,
			Name:        productResult.Name,
			Description: productResult.Description,
			Price:       float32(productResult.Price),
			Stock:       productResult.Stock,
			ImageUrl:    productResult.ImageURL,
			IsSoldOut:   productResult.IsSoldOut,
			CreatedAt:   productResult.CreatedAt,
			UpdatedAt:   productResult.UpdatedAt,
		})
	}
	return &productProto.ListProductsResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询商品列表成功",
		Products: products,
		Total:    int32(ListProductsResult.Total),
		Page:     ListProductsResult.Page,
		PageSize: ListProductsResult.PageSize,
	}, nil
}
func (p *ProductHandler) GetProductByID(ctx context.Context, req *productProto.GetProductRequest) (*productProto.GetProductResponse, error) {
	result, err := p.productService.GetProductByID(ctx, req.ProductId)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询商品详情未知错误", zap.Error(err))
			return &productProto.GetProductResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.GetProductResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换为proto响应
	return &productProto.GetProductResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "查询成功",
		Product: &productProto.Product{
			ProductId:   result.ProductID,
			MerchantId:  result.MerchantID,
			Name:        result.Name,
			Description: result.Description,
			Price:       float32(result.Price),
			Stock:       result.Stock,
			ImageUrl:    result.ImageURL,
			IsSoldOut:   result.IsSoldOut,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		},
	}, nil
}

func (p *ProductHandler) DeductStock(ctx context.Context, req *productProto.DeductStockRequest) (*productProto.CommonResponse, error) {
	param := service.DeductStockParam{
		ProductID: req.ProductId,
		Num:       req.Num,
	}
	err := p.productService.DeductStock(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("扣减库存未知错误", zap.Error(err))
			return &productProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.CommonResponse{
			Code: utils.ErrCodeSystem,
			Msg:  appError.Message,
		}, nil
	}
	return &productProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "扣减库存成功",
	}, nil
}

func (p *ProductHandler) RestoreStock(ctx context.Context, req *productProto.RestoreStockRequest) (*productProto.CommonResponse, error) {
	param := service.RestoreStockParam{
		ProductID: req.ProductId,
		Num:       req.Num,
	}
	err := p.productService.RestoreStock(ctx, param)
	if err != nil {
		var appError *utils.AppError
		ok := errors.As(err, &appError)
		if !ok {
			zap.L().Error("恢复库存未知错误", zap.Error(err))
			return &productProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &productProto.CommonResponse{
			Code: utils.ErrCodeSuccess,
			Msg:  appError.Message,
		}, nil
	}
	return &productProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "恢复库存成功",
	}, nil
}
