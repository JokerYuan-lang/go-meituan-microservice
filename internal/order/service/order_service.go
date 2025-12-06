package service

import (
	"context"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/order/client"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/order/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/order/repo/model"
	productProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/product/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 入参结构体
type CreateOrderParam struct {
	UserID             int64            `validate:"required,gt=0"`
	UserName           string           `validate:"required,min=2"`
	UserPhone          string           `validate:"required,regexp=^1[3-9]\\d{9}$"`
	MerchantID         int64            `validate:"required,gt=0"`
	Items              []OrderItemParam `validate:"required,min=1"`
	TotalAmount        float64          `validate:"required,gt=0"`
	Address            string           `validate:"required,min=5"`
	ExpectDeliveryTime string           `validate:"omitempty"`
}

type OrderItemParam struct {
	ProductID   int64   `validate:"required,gt=0"`
	ProductName string  `validate:"required,min=2"`
	Price       float64 `validate:"required,gt=0"`
	Quantity    int32   `validate:"required,gt=0"`
	TotalPrice  float64 `validate:"required,gt=0"`
}

type UpdateOrderStatusParam struct {
	OrderID  int64  `validate:"required,gt=0"`
	Status   string `validate:"required,min=2"`
	Operator string `validate:"required,min=2"`
	Remark   string `validate:"omitempty"`
}

type ListUserOrdersParam struct {
	UserID   int64  `validate:"required,gt=0"`
	Status   string `validate:"omitempty"`
	Page     int32  `validate:"required,gte=1"`
	PageSize int32  `validate:"required,gte=10,lte=100"`
}

type ListMerchantOrdersParam struct {
	MerchantID int64  `validate:"required,gt=0"`
	Status     string `validate:"omitempty"`
	Page       int32  `validate:"required,gte=1"`
	PageSize   int32  `validate:"required,gte=10,lte=100"`
}

type CancelOrderParam struct {
	OrderID int64  `validate:"required,gt=0"`
	UserID  int64  `validate:"required,gt=0"`
	Reason  string `validate:"required,min=2"`
}

// 响应结构体
type CreateOrderResult struct {
	OrderID int64  `json:"order_id"`
	OrderNo string `json:"order_no"`
}

type OrderInfoResult struct {
	OrderID            int64             `json:"order_id"`
	OrderNo            string            `json:"order_no"`
	UserID             int64             `json:"user_id"`
	UserName           string            `json:"user_name"`
	UserPhone          string            `json:"user_phone"`
	MerchantID         int64             `json:"merchant_id"`
	MerchantName       string            `json:"merchant_name"`
	Items              []OrderItemResult `json:"items"`
	TotalAmount        float64           `json:"total_amount"`
	Status             string            `json:"status"`
	Address            string            `json:"address"`
	CreateTime         string            `json:"create_time"`
	UpdateTime         string            `json:"update_time"`
	ExpectDeliveryTime string            `json:"expect_delivery_time"`
	Remark             string            `json:"remark"`
}

type OrderItemResult struct {
	ItemID      int64   `json:"item_id"`
	OrderID     int64   `json:"order_id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int32   `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}

type ListOrdersResult struct {
	Orders   []OrderInfoResult `json:"orders"`
	Total    int32             `json:"total"`
	Page     int32             `json:"page"`
	PageSize int32             `json:"page_size"`
}

// OrderService 订单业务逻辑接口
type OrderService interface {
	CreateOrder(ctx context.Context, param CreateOrderParam) (CreateOrderResult, error)
	UpdateOrderStatus(ctx context.Context, param UpdateOrderStatusParam) error
	ListUserOrders(ctx context.Context, param ListUserOrdersParam) (ListOrdersResult, error)
	ListMerchantOrders(ctx context.Context, param ListMerchantOrdersParam) (ListOrdersResult, error)
	GetOrderByID(ctx context.Context, orderID int64) (OrderInfoResult, error)
	CancelOrder(ctx context.Context, param CancelOrderParam) error
}

// orderService 实现
type orderService struct {
	orderRepo repo.OrderRepo
	validate  *validator.Validate
}

// NewOrderService 创建实例
func NewOrderService(orderRepo repo.OrderRepo) OrderService {
	return &orderService{
		orderRepo: orderRepo,
		validate:  validator.New(),
	}
}

// CreateOrder 创建订单（核心：扣减库存+事务创建订单）
func (s *orderService) CreateOrder(ctx context.Context, param CreateOrderParam) (CreateOrderResult, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("创建订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return CreateOrderResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 批量扣减商品库存（调用商品服务）
	for _, item := range param.Items {
		deductReq := &productProto.DeductStockRequest{
			ProductId: item.ProductID,
			Num:       item.Quantity,
		}
		_, err := client.ProductClient.DeductStock(ctx, deductReq)
		if err != nil {
			zap.L().Error("扣减商品库存失败", zap.Int64("product_id", item.ProductID), zap.Error(err))
			// 库存不足/商品不存在，直接返回
			return CreateOrderResult{}, utils.NewBizError("商品库存不足或不存在：" + item.ProductName)
		}
	}

	// 3. 转换为模型（订单主表）
	order := &model.Order{
		UserID:             param.UserID,
		UserName:           param.UserName,
		UserPhone:          param.UserPhone,
		MerchantID:         param.MerchantID,
		MerchantName:       "测试商家", // TODO：后续对接商家服务获取真实名称
		TotalAmount:        param.TotalAmount,
		Status:             "待接单",
		Address:            param.Address,
		ExpectDeliveryTime: param.ExpectDeliveryTime,
	}

	// 4. 转换为模型（订单项）
	var items []*model.OrderItem
	for _, item := range param.Items {
		items = append(items, &model.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			TotalPrice:  item.TotalPrice,
		})
	}

	// 5. 事务创建订单+订单项
	if err := s.orderRepo.CreateOrder(ctx, order, items); err != nil {
		// 订单创建失败，恢复库存
		for _, item := range param.Items {
			restoreReq := &productProto.RestoreStockRequest{
				ProductId: item.ProductID,
				Num:       item.Quantity,
			}
			_, _ = client.ProductClient.RestoreStock(ctx, restoreReq) // 忽略错误，仅日志
		}
		zap.L().Error("创建订单失败，已恢复库存", zap.Int64("user_id", param.UserID), zap.Error(err))
		return CreateOrderResult{}, err
	}

	// 6. 组装结果
	result := CreateOrderResult{
		OrderID: order.OrderID,
		OrderNo: order.OrderNo,
	}

	zap.L().Info("创建订单成功", zap.Int64("order_id", order.OrderID), zap.String("order_no", order.OrderNo), zap.Int64("user_id", param.UserID))
	return result, nil
}

// UpdateOrderStatus 更新订单状态
func (s *orderService) UpdateOrderStatus(ctx context.Context, param UpdateOrderStatusParam) error {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("更新订单状态参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	//// 2. 校验状态合法性
	//validStatus := []string{"待接单", "已接单", "待配送", "已完成", "已取消", "已拒单"}
	//if !utils.ContainsString(validStatus, param.Status) {
	//	return utils.NewParamError("订单状态不合法")
	//}

	// 3. 调用Repo更新状态
	return s.orderRepo.UpdateOrderStatus(ctx, param.OrderID, param.Status, param.Remark)
}

// ListUserOrders 查询用户订单列表
func (s *orderService) ListUserOrders(ctx context.Context, param ListUserOrdersParam) (ListOrdersResult, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询用户订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return ListOrdersResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用Repo查询订单
	orders, total, err := s.orderRepo.ListUserOrders(ctx, param.UserID, param.Status, param.Page, param.PageSize)
	if err != nil {
		return ListOrdersResult{}, err
	}

	// 3. 批量查询订单项
	var resultOrders []OrderInfoResult
	for _, o := range orders {
		// 查询订单项
		items, err := s.orderRepo.GetOrderItems(ctx, o.OrderID)
		if err != nil {
			zap.L().Warn("查询订单项失败，跳过该订单", zap.Int64("order_id", o.OrderID), zap.Error(err))
			continue
		}

		// 转换订单项
		var itemResults []OrderItemResult
		for _, item := range items {
			itemResults = append(itemResults, OrderItemResult{
				ItemID:      item.ItemID,
				OrderID:     item.OrderID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
				TotalPrice:  item.TotalPrice,
			})
		}

		// 转换订单
		resultOrders = append(resultOrders, OrderInfoResult{
			OrderID:            o.OrderID,
			OrderNo:            o.OrderNo,
			UserID:             o.UserID,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			MerchantID:         o.MerchantID,
			MerchantName:       o.MerchantName,
			Items:              itemResults,
			TotalAmount:        o.TotalAmount,
			Status:             o.Status,
			Address:            o.Address,
			CreateTime:         o.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:         o.UpdateTime.Format("2006-01-02 15:04:05"),
			ExpectDeliveryTime: o.ExpectDeliveryTime,
			Remark:             o.Remark,
		})
	}

	// 4. 组装结果
	result := ListOrdersResult{
		Orders:   resultOrders,
		Total:    int32(total),
		Page:     param.Page,
		PageSize: param.PageSize,
	}

	return result, nil
}

// ListMerchantOrders 查询商家订单列表
func (s *orderService) ListMerchantOrders(ctx context.Context, param ListMerchantOrdersParam) (ListOrdersResult, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询商家订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return ListOrdersResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用Repo查询订单
	orders, total, err := s.orderRepo.ListMerchantOrders(ctx, param.MerchantID, param.Status, param.Page, param.PageSize)
	if err != nil {
		return ListOrdersResult{}, err
	}

	// 3. 批量查询订单项
	var resultOrders []OrderInfoResult
	for _, o := range orders {
		// 查询订单项
		items, err := s.orderRepo.GetOrderItems(ctx, o.OrderID)
		if err != nil {
			zap.L().Warn("查询订单项失败，跳过该订单", zap.Int64("order_id", o.OrderID), zap.Error(err))
			continue
		}

		// 转换订单项
		var itemResults []OrderItemResult
		for _, item := range items {
			itemResults = append(itemResults, OrderItemResult{
				ItemID:      item.ItemID,
				OrderID:     item.OrderID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
				TotalPrice:  item.TotalPrice,
			})
		}

		// 转换订单
		resultOrders = append(resultOrders, OrderInfoResult{
			OrderID:            o.OrderID,
			OrderNo:            o.OrderNo,
			UserID:             o.UserID,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			MerchantID:         o.MerchantID,
			MerchantName:       o.MerchantName,
			Items:              itemResults,
			TotalAmount:        o.TotalAmount,
			Status:             o.Status,
			Address:            o.Address,
			CreateTime:         o.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:         o.UpdateTime.Format("2006-01-02 15:04:05"),
			ExpectDeliveryTime: o.ExpectDeliveryTime,
			Remark:             o.Remark,
		})
	}

	// 4. 组装结果
	result := ListOrdersResult{
		Orders:   resultOrders,
		Total:    int32(total),
		Page:     param.Page,
		PageSize: param.PageSize,
	}

	return result, nil
}

// GetOrderByID 查询订单详情
func (s *orderService) GetOrderByID(ctx context.Context, orderID int64) (OrderInfoResult, error) {
	// 1. 参数校验
	if orderID <= 0 {
		return OrderInfoResult{}, utils.NewParamError("订单ID不能为空且大于0")
	}

	// 2. 查询订单主表
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return OrderInfoResult{}, err
	}

	// 3. 查询订单项
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		zap.L().Warn("查询订单项失败", zap.Int64("order_id", orderID), zap.Error(err))
		items = []*model.OrderItem{} // 空列表，不影响主信息
	}

	// 4. 转换订单项
	var itemResults []OrderItemResult
	for _, item := range items {
		itemResults = append(itemResults, OrderItemResult{
			ItemID:      item.ItemID,
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			TotalPrice:  item.TotalPrice,
		})
	}

	// 5. 组装结果
	result := OrderInfoResult{
		OrderID:            order.OrderID,
		OrderNo:            order.OrderNo,
		UserID:             order.UserID,
		UserName:           order.UserName,
		UserPhone:          order.UserPhone,
		MerchantID:         order.MerchantID,
		MerchantName:       order.MerchantName,
		Items:              itemResults,
		TotalAmount:        order.TotalAmount,
		Status:             order.Status,
		Address:            order.Address,
		CreateTime:         order.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:         order.UpdateTime.Format("2006-01-02 15:04:05"),
		ExpectDeliveryTime: order.ExpectDeliveryTime,
		Remark:             order.Remark,
	}

	return result, nil
}

// CancelOrder 取消订单（恢复库存+更新状态）
func (s *orderService) CancelOrder(ctx context.Context, param CancelOrderParam) error {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("取消订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 查询订单详情（校验状态：仅待接单/已拒单可取消）
	order, err := s.orderRepo.GetOrderByID(ctx, param.OrderID)
	if err != nil {
		return err
	}
	if order.Status != "待接单" && order.Status != "已拒单" {
		return utils.NewBizError("仅待接单/已拒单的订单可取消")
	}

	// 3. 查询订单项，恢复库存
	items, err := s.orderRepo.GetOrderItems(ctx, param.OrderID)
	if err != nil {
		zap.L().Warn("查询订单项失败，跳过库存恢复", zap.Int64("order_id", param.OrderID), zap.Error(err))
	} else {
		for _, item := range items {
			restoreReq := &productProto.RestoreStockRequest{
				ProductId: item.ProductID,
				Num:       item.Quantity,
			}
			_, _ = client.ProductClient.RestoreStock(ctx, restoreReq) // 忽略错误，仅日志
		}
	}

	// 4. 调用Repo取消订单
	return s.orderRepo.CancelOrder(ctx, param.OrderID, param.UserID, param.Reason)
}
