package handler

import (
	"context"
	"errors"

	orderProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/order/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/order/service"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
)

// OrderHandler 订单gRPC接口实现
type OrderHandler struct {
	orderProto.UnimplementedOrderServiceServer
	orderService service.OrderService
}

// NewOrderHandler 创建实例
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderProto.CreateOrderRequest) (*orderProto.CreateOrderResponse, error) {
	// 1. 转换订单项
	var items []service.OrderItemParam
	for _, item := range req.Items {
		items = append(items, service.OrderItemParam{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Price:       float64(item.Price),
			Quantity:    item.Quantity,
			TotalPrice:  float64(item.TotalPrice),
		})
	}

	// 2. proto → service参数
	param := service.CreateOrderParam{
		UserID:             req.UserId,
		UserName:           req.UserName,
		UserPhone:          req.UserPhone,
		MerchantID:         req.MerchantId,
		Items:              items,
		TotalAmount:        float64(req.TotalAmount),
		Address:            req.Address,
		ExpectDeliveryTime: req.ExpectDeliveryTime,
	}

	// 3. 调用service
	result, err := h.orderService.CreateOrder(ctx, param)
	if err != nil {
		appErr, ok := err.(*utils.AppError)
		if !ok {
			zap.L().Error("创建订单未知错误", zap.Error(err))
			return &orderProto.CreateOrderResponse{
				Code:    utils.ErrCodeSystem,
				Msg:     "系统错误",
				OrderId: 0,
				OrderNo: "",
			}, nil
		}
		return &orderProto.CreateOrderResponse{
			Code:    int32(appErr.Code),
			Msg:     appErr.Message,
			OrderId: 0,
			OrderNo: "",
		}, nil
	}

	// 4. 响应转换
	return &orderProto.CreateOrderResponse{
		Code:    utils.ErrCodeSuccess,
		Msg:     "创建订单成功",
		OrderId: result.OrderID,
		OrderNo: result.OrderNo,
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *orderProto.UpdateOrderStatusRequest) (*orderProto.CommonResponse, error) {
	// proto → service参数
	param := service.UpdateOrderStatusParam{
		OrderID:  req.OrderId,
		Status:   req.Status,
		Operator: req.Operator,
		Remark:   req.Remark,
	}

	// 调用service
	err := h.orderService.UpdateOrderStatus(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("更新订单状态未知错误", zap.Error(err), zap.Int64("order_id", req.OrderId))
			return &orderProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &orderProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &orderProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新订单状态成功",
	}, nil
}

// ListUserOrders 查询用户订单列表
func (h *OrderHandler) ListUserOrders(ctx context.Context, req *orderProto.ListUserOrdersRequest) (*orderProto.ListUserOrdersResponse, error) {
	// proto → service参数
	param := service.ListUserOrdersParam{
		UserID:   req.UserId,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 调用service
	result, err := h.orderService.ListUserOrders(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询用户订单未知错误", zap.Error(err), zap.Int64("user_id", req.UserId))
			return &orderProto.ListUserOrdersResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &orderProto.ListUserOrdersResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换为proto响应
	var protoOrders []*orderProto.Order
	for _, o := range result.Orders {
		// 转换订单项
		var protoItems []*orderProto.OrderItem
		for _, item := range o.Items {
			protoItems = append(protoItems, &orderProto.OrderItem{
				ItemId:      item.ItemID,
				OrderId:     item.OrderID,
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Price:       float32(item.Price),
				Quantity:    item.Quantity,
				TotalPrice:  float32(item.TotalPrice),
			})
		}

		// 转换订单
		protoOrders = append(protoOrders, &orderProto.Order{
			OrderId:            o.OrderID,
			OrderNo:            o.OrderNo,
			UserId:             o.UserID,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			MerchantId:         o.MerchantID,
			MerchantName:       o.MerchantName,
			Items:              protoItems,
			TotalAmount:        float32(o.TotalAmount),
			Status:             o.Status,
			Address:            o.Address,
			CreateTime:         o.CreateTime,
			UpdateTime:         o.UpdateTime,
			ExpectDeliveryTime: o.ExpectDeliveryTime,
			Remark:             o.Remark,
		})
	}

	return &orderProto.ListUserOrdersResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Orders:   protoOrders,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// ListMerchantOrders 查询商家订单列表
func (h *OrderHandler) ListMerchantOrders(ctx context.Context, req *orderProto.ListMerchantOrdersRequest) (*orderProto.ListMerchantOrdersResponse, error) {
	// proto → service参数
	param := service.ListMerchantOrdersParam{
		MerchantID: req.MerchantId,
		Status:     req.Status,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 调用service
	result, err := h.orderService.ListMerchantOrders(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询商家订单未知错误", zap.Error(err), zap.Int64("merchant_id", req.MerchantId))
			return &orderProto.ListMerchantOrdersResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &orderProto.ListMerchantOrdersResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换为proto响应
	var protoOrders []*orderProto.Order
	for _, o := range result.Orders {
		// 转换订单项
		var protoItems []*orderProto.OrderItem
		for _, item := range o.Items {
			protoItems = append(protoItems, &orderProto.OrderItem{
				ItemId:      item.ItemID,
				OrderId:     item.OrderID,
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Price:       float32(item.Price),
				Quantity:    item.Quantity,
				TotalPrice:  float32(item.TotalPrice),
			})
		}

		// 转换订单
		protoOrders = append(protoOrders, &orderProto.Order{
			OrderId:            o.OrderID,
			OrderNo:            o.OrderNo,
			UserId:             o.UserID,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			MerchantId:         o.MerchantID,
			MerchantName:       o.MerchantName,
			Items:              protoItems,
			TotalAmount:        float32(o.TotalAmount),
			Status:             o.Status,
			Address:            o.Address,
			CreateTime:         o.CreateTime,
			UpdateTime:         o.UpdateTime,
			ExpectDeliveryTime: o.ExpectDeliveryTime,
			Remark:             o.Remark,
		})
	}

	return &orderProto.ListMerchantOrdersResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "查询成功",
		Orders:   protoOrders,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// GetOrderByID 查询订单详情
func (h *OrderHandler) GetOrderByID(ctx context.Context, req *orderProto.GetOrderRequest) (*orderProto.GetOrderResponse, error) {
	// 调用service
	result, err := h.orderService.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("查询订单详情未知错误", zap.Error(err), zap.Int64("order_id", req.OrderId))
			return &orderProto.GetOrderResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &orderProto.GetOrderResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 转换订单项
	var protoItems []*orderProto.OrderItem
	for _, item := range result.Items {
		protoItems = append(protoItems, &orderProto.OrderItem{
			ItemId:      item.ItemID,
			OrderId:     item.OrderID,
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       float32(item.Price),
			Quantity:    item.Quantity,
			TotalPrice:  float32(item.TotalPrice),
		})
	}

	// 转换订单
	protoOrder := &orderProto.Order{
		OrderId:            result.OrderID,
		OrderNo:            result.OrderNo,
		UserId:             result.UserID,
		UserName:           result.UserName,
		UserPhone:          result.UserPhone,
		MerchantId:         result.MerchantID,
		MerchantName:       result.MerchantName,
		Items:              protoItems,
		TotalAmount:        float32(result.TotalAmount),
		Status:             result.Status,
		Address:            result.Address,
		CreateTime:         result.CreateTime,
		UpdateTime:         result.UpdateTime,
		ExpectDeliveryTime: result.ExpectDeliveryTime,
		Remark:             result.Remark,
	}

	return &orderProto.GetOrderResponse{
		Code:  utils.ErrCodeSuccess,
		Msg:   "查询成功",
		Order: protoOrder,
	}, nil
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(ctx context.Context, req *orderProto.CancelOrderRequest) (*orderProto.CommonResponse, error) {
	// proto → service参数
	param := service.CancelOrderParam{
		OrderID: req.OrderId,
		UserID:  req.UserId,
		Reason:  req.Reason,
	}

	// 调用service
	err := h.orderService.CancelOrder(ctx, param)
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("取消订单未知错误", zap.Error(err), zap.Int64("order_id", req.OrderId))
			return &orderProto.CommonResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &orderProto.CommonResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	return &orderProto.CommonResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "取消订单成功",
	}, nil
}
