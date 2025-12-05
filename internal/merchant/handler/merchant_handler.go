package handler

import (
	"context"
	"errors"

	merchantProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/service"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
)

// MerchantHandler 商家gRPC接口实现
type MerchantHandler struct {
	merchantProto.UnimplementedMerchantServiceServer
	merchantService service.MerchantService
}

// NewMerchantHandler 创建实例
func NewMerchantHandler(merchantService service.MerchantService) *MerchantHandler {
	return &MerchantHandler{
		merchantService: merchantService,
	}
}

// MerchantRegister 商家入驻
func (h *MerchantHandler) MerchantRegister(ctx context.Context, req *merchantProto.MerchantRegisterRequest) (*merchantProto.MerchantRegisterResponse, error) {
	// proto → service参数
	param := service.MerchantRegisterParam{
		Name:          req.Name,
		Phone:         req.Phone,
		Password:      req.Password,
		Address:       req.Address,
		Logo:          req.Logo,
		BusinessHours: req.BusinessHours,
	}

	// 调用service
	merchantID, token, err := h.merchantService.MerchantRegister(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("商家入驻未知错误", zap.Error(err))
			return &merchantProto.MerchantRegisterResponse{
				Code:       utils.ErrCodeSystem,
				Msg:        "系统错误",
				MerchantId: 0,
				Token:      "",
			}, nil
		}
		return &merchantProto.MerchantRegisterResponse{
			Code:       int32(appErr.Code),
			Msg:        appErr.Message,
			MerchantId: 0,
			Token:      "",
		}, nil
	}

	// 响应转换
	return &merchantProto.MerchantRegisterResponse{
		Code:       utils.ErrCodeSuccess,
		Msg:        "入驻成功",
		MerchantId: merchantID,
		Token:      token,
	}, nil
}

// MerchantLogin 商家登录
func (h *MerchantHandler) MerchantLogin(ctx context.Context, req *merchantProto.MerchantLoginRequest) (*merchantProto.MerchantLoginResponse, error) {
	// proto → service参数
	param := service.MerchantLoginParam{
		Phone:    req.Phone,
		Password: req.Password,
	}

	// 调用service
	result, err := h.merchantService.MerchantLogin(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("商家登录未知错误", zap.Error(err))
			return &merchantProto.MerchantLoginResponse{
				Code:       utils.ErrCodeSystem,
				Msg:        "系统错误",
				MerchantId: 0,
				Name:       "",
				Token:      "",
			}, nil
		}
		return &merchantProto.MerchantLoginResponse{
			Code:       int32(appErr.Code),
			Msg:        appErr.Message,
			MerchantId: 0,
			Name:       "",
			Token:      "",
		}, nil
	}

	// 响应转换
	return &merchantProto.MerchantLoginResponse{
		Code:       utils.ErrCodeSuccess,
		Msg:        "登录成功",
		MerchantId: result.MerchantID,
		Name:       result.Name,
		Token:      result.Token,
	}, nil
}

// GetMerchantInfo 获取商家信息
func (h *MerchantHandler) GetMerchantInfo(ctx context.Context, req *merchantProto.GetMerchantInfoRequest) (*merchantProto.GetMerchantInfoResponse, error) {
	// 调用service
	result, err := h.merchantService.GetMerchantInfo(ctx, req.MerchantId)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询商家信息未知错误", zap.Error(err), zap.Int64("merchant_id", req.MerchantId))
			return &merchantProto.GetMerchantInfoResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &merchantProto.GetMerchantInfoResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换为proto响应
	merchant := &merchantProto.Merchant{
		MerchantId:    result.MerchantID,
		Name:          result.Name,
		Phone:         result.Phone,
		Address:       result.Address,
		Logo:          result.Logo,
		BusinessHours: result.BusinessHours,
		Score:         float32(result.Score),
		OrderCount:    result.OrderCount,
		IsOpen:        result.IsOpen,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	}

	return &merchantProto.GetMerchantInfoResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Merchant: merchant,
	}, nil
}

// UpdateMerchantInfo 更新商家信息
func (h *MerchantHandler) UpdateMerchantInfo(ctx context.Context, req *merchantProto.UpdateMerchantInfoRequest) (*merchantProto.CommonResponse, error) {
	// proto → service参数
	param := service.UpdateMerchantInfoParam{
		MerchantID:    req.MerchantId,
		Name:          req.Name,
		Phone:         req.Phone,
		Address:       req.Address,
		Logo:          req.Logo,
		BusinessHours: req.BusinessHours,
		IsOpen:        req.IsOpen,
	}

	// 调用service
	err := h.merchantService.UpdateMerchantInfo(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("更新商家信息未知错误", zap.Error(err), zap.Int64("merchant_id", req.MerchantId))
			return &merchantProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &merchantProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &merchantProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新成功",
	}, nil
}

// AcceptOrder 商家接单
func (h *MerchantHandler) AcceptOrder(ctx context.Context, req *merchantProto.AcceptOrderRequest) (*merchantProto.CommonResponse, error) {
	// proto → service参数
	param := service.AcceptOrderParam{
		OrderID:    req.OrderId,
		MerchantID: req.MerchantId,
	}

	// 调用service
	err := h.merchantService.AcceptOrder(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("商家接单未知错误", zap.Error(err), zap.Any("req", req))
			return &merchantProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &merchantProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &merchantProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "接单成功",
	}, nil
}

// RejectOrder 商家拒单
func (h *MerchantHandler) RejectOrder(ctx context.Context, req *merchantProto.RejectOrderRequest) (*merchantProto.CommonResponse, error) {
	// proto → service参数
	param := service.RejectOrderParam{
		OrderID:    req.OrderId,
		MerchantID: req.MerchantId,
		Reason:     req.Reason,
	}

	// 调用service
	err := h.merchantService.RejectOrder(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("商家拒单未知错误", zap.Error(err), zap.Any("req", req))
			return &merchantProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &merchantProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &merchantProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "拒单成功",
	}, nil
}

// ListMerchantOrders 查询商家订单列表
func (h *MerchantHandler) ListMerchantOrders(ctx context.Context, req *merchantProto.ListMerchantOrdersRequest) (*merchantProto.ListMerchantOrdersResponse, error) {
	// proto → service参数
	param := service.ListMerchantOrdersParam{
		MerchantID: req.MerchantId,
		Status:     req.Status,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 调用service
	result, err := h.merchantService.ListMerchantOrders(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询商家订单未知错误", zap.Error(err), zap.Any("req", req))
			return &merchantProto.ListMerchantOrdersResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &merchantProto.ListMerchantOrdersResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换为proto响应
	var protoOrders []*merchantProto.MerchantOrder
	for _, o := range result.Orders {
		protoOrders = append(protoOrders, &merchantProto.MerchantOrder{
			OrderId:            o.OrderID,
			UserId:             o.UserID,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			TotalAmount:        float32(o.TotalAmount),
			Status:             o.Status,
			CreateTime:         o.CreateTime,
			ExpectDeliveryTime: o.ExpectDeliveryTime,
		})
	}

	return &merchantProto.ListMerchantOrdersResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Orders:   protoOrders,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
