package handler

import (
	"context"
	"errors"

	riderProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/service"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
)

// RiderHandler 骑手gRPC接口实现
type RiderHandler struct {
	riderProto.UnimplementedRiderServiceServer
	riderService service.RiderService
}

// NewRiderHandler 创建实例
func NewRiderHandler(riderService service.RiderService) *RiderHandler {
	return &RiderHandler{
		riderService: riderService,
	}
}

// RiderRegister 骑手注册
func (h *RiderHandler) RiderRegister(ctx context.Context, req *riderProto.RiderRegisterRequest) (*riderProto.RiderRegisterResponse, error) {
	// 转换参数
	param := service.RiderRegisterParam{
		Name:     req.Name,
		Phone:    req.Phone,
		Password: req.Password,
		Avatar:   req.Avatar,
	}

	// 调用service
	riderID, token, err := h.riderService.RiderRegister(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("骑手注册未知错误", zap.Error(err))
			return &riderProto.RiderRegisterResponse{
				Code:    utils.ErrCodeSystem,
				Msg:     "系统错误",
				RiderId: 0,
				Token:   "",
			}, nil
		}
		return &riderProto.RiderRegisterResponse{
			Code:    int32(appErr.Code),
			Msg:     appErr.Message,
			RiderId: 0,
			Token:   "",
		}, nil
	}

	return &riderProto.RiderRegisterResponse{
		Code:    utils.ErrCodeSuccess,
		Msg:     "注册成功",
		RiderId: riderID,
		Token:   token,
	}, nil
}

// RiderLogin 骑手登录
func (h *RiderHandler) RiderLogin(ctx context.Context, req *riderProto.RiderLoginRequest) (*riderProto.RiderLoginResponse, error) {
	// 转换参数
	param := service.RiderLoginParam{
		Phone:    req.Phone,
		Password: req.Password,
	}

	// 调用service
	result, err := h.riderService.RiderLogin(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("骑手登录未知错误", zap.Error(err))
			return &riderProto.RiderLoginResponse{
				Code:    utils.ErrCodeSystem,
				Msg:     "系统错误",
				RiderId: 0,
				Name:    "",
				Token:   "",
			}, nil
		}
		return &riderProto.RiderLoginResponse{
			Code:    int32(appErr.Code),
			Msg:     appErr.Message,
			RiderId: 0,
			Name:    "",
			Token:   "",
		}, nil
	}

	return &riderProto.RiderLoginResponse{
		Code:    utils.ErrCodeSuccess,
		Msg:     "登录成功",
		RiderId: result.RiderID,
		Name:    result.Name,
		Token:   result.Token,
	}, nil
}

// GetRiderInfo 获取骑手信息
func (h *RiderHandler) GetRiderInfo(ctx context.Context, req *riderProto.GetRiderInfoRequest) (*riderProto.GetRiderInfoResponse, error) {
	// 调用service
	result, err := h.riderService.GetRiderInfo(ctx, req.RiderId)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询骑手信息未知错误", zap.Error(err), zap.Int64("rider_id", req.RiderId))
			return &riderProto.GetRiderInfoResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &riderProto.GetRiderInfoResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换响应
	riderPro := &riderProto.Rider{
		RiderId:    result.RiderID,
		Name:       result.Name,
		Phone:      result.Phone,
		Avatar:     result.Avatar,
		Score:      float32(result.Score),
		OrderCount: result.OrderCount,
		Status:     result.Status,
		CreateTime: result.CreatedAt,
		UpdateTime: result.UpdatedAt,
	}

	return &riderProto.GetRiderInfoResponse{
		Code:  utils.ErrCodeSuccess,
		Msg:   "查询成功",
		Rider: riderPro,
	}, nil
}

// AcceptOrder 骑手接单
func (h *RiderHandler) AcceptOrder(ctx context.Context, req *riderProto.AcceptOrderRequest) (*riderProto.CommonResponse, error) {
	// 转换参数
	param := service.AcceptOrderParam{
		OrderID: req.OrderId,
		RiderID: req.RiderId,
	}

	// 调用service
	err := h.riderService.AcceptOrder(ctx, param)
	if err != nil {
		appErr, ok := err.(*utils.AppError)
		if !ok {
			zap.L().Error("骑手接单未知错误", zap.Error(err), zap.Any("req", req))
			return &riderProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &riderProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &riderProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "接单成功",
	}, nil
}

// UpdateDeliveryStatus 更新配送状态
func (h *RiderHandler) UpdateDeliveryStatus(ctx context.Context, req *riderProto.UpdateDeliveryStatusRequest) (*riderProto.CommonResponse, error) {
	// 转换参数
	param := service.UpdateDeliveryStatusParam{
		OrderID:        req.OrderId,
		RiderID:        req.RiderId,
		DeliveryStatus: req.DeliveryStatus,
	}

	// 调用service
	err := h.riderService.UpdateDeliveryStatus(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("更新配送状态未知错误", zap.Error(err), zap.Any("req", req))
			return &riderProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &riderProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &riderProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新配送状态成功",
	}, nil
}

// ListPendingOrders 查询待接订单列表
func (h *RiderHandler) ListPendingOrders(ctx context.Context, req *riderProto.ListPendingOrdersRequest) (*riderProto.ListPendingOrdersResponse, error) {
	// 转换参数
	param := service.ListPendingOrdersParam{
		Area:     req.Area,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 调用service
	result, err := h.riderService.ListPendingOrders(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询待接订单未知错误", zap.Error(err), zap.Any("req", req))
			return &riderProto.ListPendingOrdersResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &riderProto.ListPendingOrdersResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换响应
	var protoOrders []*riderProto.DeliveryOrder
	for _, o := range result.Orders {
		protoOrders = append(protoOrders, &riderProto.DeliveryOrder{
			OrderId:        o.OrderID,
			OrderNo:        o.OrderNo,
			RiderId:        o.RiderID,
			RiderName:      o.RiderName,
			MerchantId:     o.MerchantID,
			MerchantName:   o.MerchantName,
			Address:        o.Address,
			TotalAmount:    float32(o.TotalAmount),
			DeliveryStatus: o.DeliveryStatus,
			AcceptTime:     o.AcceptTime,
			PickupTime:     o.PickupTime,
			CompleteTime:   o.CompleteTime,
		})
	}

	return &riderProto.ListPendingOrdersResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Orders:   protoOrders,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// ListRiderOrders 查询骑手配送订单
func (h *RiderHandler) ListRiderOrders(ctx context.Context, req *riderProto.ListRiderOrdersRequest) (*riderProto.ListRiderOrdersResponse, error) {
	// 转换参数
	param := service.ListRiderOrdersParam{
		RiderID:        req.RiderId,
		DeliveryStatus: req.DeliveryStatus,
		Page:           req.Page,
		PageSize:       req.PageSize,
	}

	// 调用service
	result, err := h.riderService.ListRiderOrders(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询骑手订单未知错误", zap.Error(err), zap.Any("req", req))
			return &riderProto.ListRiderOrdersResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &riderProto.ListRiderOrdersResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换响应
	var protoOrders []*riderProto.DeliveryOrder
	for _, o := range result.Orders {
		protoOrders = append(protoOrders, &riderProto.DeliveryOrder{
			OrderId:        o.OrderID,
			OrderNo:        o.OrderNo,
			RiderId:        o.RiderID,
			RiderName:      o.RiderName,
			MerchantId:     o.MerchantID,
			MerchantName:   o.MerchantName,
			Address:        o.Address,
			TotalAmount:    float32(o.TotalAmount),
			DeliveryStatus: o.DeliveryStatus,
			AcceptTime:     o.AcceptTime,
			PickupTime:     o.PickupTime,
			CompleteTime:   o.CompleteTime,
		})
	}

	return &riderProto.ListRiderOrdersResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Orders:   protoOrders,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
