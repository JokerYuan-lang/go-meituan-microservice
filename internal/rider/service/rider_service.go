package service

import (
	"context"
	"strconv"
	"time"

	orderProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/order/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/client"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 入参结构体
type RiderRegisterParam struct {
	Name     string `validate:"required,min=2"`
	Phone    string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Password string `validate:"required,min=6"`
	Avatar   string `validate:"required,url"`
}

type RiderLoginParam struct {
	Phone    string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Password string `validate:"required,min=6"`
}

type AcceptOrderParam struct {
	OrderID int64 `validate:"required,gt=0"`
	RiderID int64 `validate:"required,gt=0"`
}

type UpdateDeliveryStatusParam struct {
	OrderID        int64  `validate:"required,gt=0"`
	RiderID        int64  `validate:"required,gt=0"`
	DeliveryStatus string `validate:"required,min=2"`
}

type ListPendingOrdersParam struct {
	Area     string `validate:"omitempty"`
	Page     int32  `validate:"required,gte=1"`
	PageSize int32  `validate:"required,gte=10,lte=100"`
}

type ListRiderOrdersParam struct {
	RiderID        int64  `validate:"required,gt=0"`
	DeliveryStatus string `validate:"omitempty"`
	Page           int32  `validate:"required,gte=1"`
	PageSize       int32  `validate:"required,gte=10,lte=100"`
}

// 响应结构体
type RiderLoginResult struct {
	RiderID int64  `json:"rider_id"`
	Name    string `json:"name"`
	Token   string `json:"token"`
}

type RiderInfoResult struct {
	RiderID    int64   `json:"rider_id"`
	Name       string  `json:"name"`
	Phone      string  `json:"phone"`
	Avatar     string  `json:"avatar"`
	Score      float64 `json:"score"`
	OrderCount int32   `json:"order_count"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type DeliveryOrderResult struct {
	OrderID        int64   `json:"order_id"`
	OrderNo        string  `json:"order_no"`
	RiderID        int64   `json:"rider_id"`
	RiderName      string  `json:"rider_name"`
	MerchantID     int64   `json:"merchant_id"`
	MerchantName   string  `json:"merchant_name"`
	Address        string  `json:"address"`
	TotalAmount    float64 `json:"total_amount"`
	DeliveryStatus string  `json:"delivery_status"`
	AcceptTime     string  `json:"accept_time"`
	PickupTime     string  `json:"pickup_time"`
	CompleteTime   string  `json:"complete_time"`
}

type ListOrdersResult struct {
	Orders   []DeliveryOrderResult `json:"orders"`
	Total    int32                 `json:"total"`
	Page     int32                 `json:"page"`
	PageSize int32                 `json:"page_size"`
}

// RiderService 骑手业务逻辑接口
type RiderService interface {
	RiderRegister(ctx context.Context, param RiderRegisterParam) (int64, string, error)
	RiderLogin(ctx context.Context, param RiderLoginParam) (RiderLoginResult, error)
	GetRiderInfo(ctx context.Context, riderID int64) (RiderInfoResult, error)
	AcceptOrder(ctx context.Context, param AcceptOrderParam) error
	UpdateDeliveryStatus(ctx context.Context, param UpdateDeliveryStatusParam) error
	ListPendingOrders(ctx context.Context, param ListPendingOrdersParam) (ListOrdersResult, error)
	ListRiderOrders(ctx context.Context, param ListRiderOrdersParam) (ListOrdersResult, error)
}

// riderService 实现
type riderService struct {
	riderRepo repo.RiderRepo
	validate  *validator.Validate
}

// NewRiderService 创建实例
func NewRiderService(riderRepo repo.RiderRepo) RiderService {
	return &riderService{
		riderRepo: riderRepo,
		validate:  validator.New(),
	}
}

// RiderRegister 骑手注册
func (s *riderService) RiderRegister(ctx context.Context, param RiderRegisterParam) (int64, string, error) {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("骑手注册参数校验失败", zap.Any("param", param), zap.Error(err))
		return 0, "", utils.NewParamError("参数错误：" + err.Error())
	}

	// 校验手机号是否已注册
	existRider, err := s.riderRepo.GetRiderByPhone(ctx, param.Phone)
	if err != nil {
		return 0, "", err
	}
	if existRider != nil {
		return 0, "", utils.NewBizError("手机号已注册")
	}

	// 转换为模型
	rider := &model.Rider{
		Name:     param.Name,
		Phone:    param.Phone,
		Password: param.Password,
		Avatar:   param.Avatar,
		Status:   "在线",
	}

	// 创建骑手
	if err := s.riderRepo.CreateRider(ctx, rider); err != nil {
		return 0, "", err
	}

	// 生成Token
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(rider.RiderID, 10),
		Username: rider.Name,
		Phone:    rider.Phone,
		Role:     "rider", // 骑手角色
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成骑手Token失败", zap.Int64("rider_id", rider.RiderID), zap.Error(err))
		return rider.RiderID, "", utils.NewSystemError("注册成功，但生成Token失败")
	}

	zap.L().Info("骑手注册成功", zap.Int64("rider_id", rider.RiderID), zap.String("phone", param.Phone))
	return rider.RiderID, token, nil
}

// RiderLogin 骑手登录
func (s *riderService) RiderLogin(ctx context.Context, param RiderLoginParam) (RiderLoginResult, error) {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("骑手登录参数校验失败", zap.Any("param", param), zap.Error(err))
		return RiderLoginResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 查询骑手
	rider, err := s.riderRepo.GetRiderByPhone(ctx, param.Phone)
	if err != nil {
		return RiderLoginResult{}, err
	}
	if rider == nil {
		return RiderLoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 验证密码
	if !utils.CheckPasswordHash(param.Password, rider.Password) {
		return RiderLoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 生成Token
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(rider.RiderID, 10),
		Username: rider.Name,
		Phone:    rider.Phone,
		Role:     "rider",
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成骑手登录Token失败", zap.Int64("rider_id", rider.RiderID), zap.Error(err))
		return RiderLoginResult{}, utils.NewSystemError("登录失败，生成Token失败")
	}

	// 组装结果
	result := RiderLoginResult{
		RiderID: rider.RiderID,
		Name:    rider.Name,
		Token:   token,
	}

	zap.L().Info("骑手登录成功", zap.Int64("rider_id", rider.RiderID), zap.String("phone", param.Phone))
	return result, nil
}

// GetRiderInfo 获取骑手信息
func (s *riderService) GetRiderInfo(ctx context.Context, riderID int64) (RiderInfoResult, error) {
	// 参数校验
	if riderID <= 0 {
		return RiderInfoResult{}, utils.NewParamError("骑手ID不能为空且大于0")
	}

	// 查询骑手
	rider, err := s.riderRepo.GetRiderByID(ctx, riderID)
	if err != nil {
		return RiderInfoResult{}, err
	}

	// 组装结果
	result := RiderInfoResult{
		RiderID:    rider.RiderID,
		Name:       rider.Name,
		Phone:      rider.Phone,
		Avatar:     rider.Avatar,
		Score:      rider.Score,
		OrderCount: rider.OrderCount,
		Status:     rider.Status,
		CreatedAt:  rider.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  rider.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// AcceptOrder 骑手接单（关联订单+更新订单服务状态）
func (s *riderService) AcceptOrder(ctx context.Context, param AcceptOrderParam) error {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("骑手接单参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 校验骑手是否在线
	rider, err := s.riderRepo.GetRiderByID(ctx, param.RiderID)
	if err != nil {
		return err
	}
	if rider.Status != "在线" {
		return utils.NewBizError("骑手当前离线，无法接单")
	}

	// 1. 更新配送订单状态为「待取餐」
	now := time.Now().Format("2006-01-02 15:04:05")
	if err := s.riderRepo.UpdateDeliveryOrder(ctx, param.OrderID, param.RiderID, "待取餐", now); err != nil {
		return err
	}

	// 2. 调用订单服务更新订单状态为「待配送」
	updateStatusReq := &orderProto.UpdateOrderStatusRequest{
		OrderId:  param.OrderID,
		Status:   "待配送",
		Operator: "rider_" + strconv.FormatInt(param.RiderID, 10),
	}
	_, err = client.OrderClient.UpdateOrderStatus(ctx, updateStatusReq)
	if err != nil {
		zap.L().Error("调用订单服务更新状态失败", zap.Int64("order_id", param.OrderID), zap.Error(err))
		return utils.NewSystemError("接单失败，订单服务异常")
	}

	// 3. 更新骑手订单数+1
	if err := s.riderRepo.UpdateOrderCount(ctx, param.RiderID, 1); err != nil {
		zap.L().Warn("更新骑手订单数失败", zap.Int64("rider_id", param.RiderID), zap.Error(err))
	}

	zap.L().Info("骑手接单成功", zap.Int64("order_id", param.OrderID), zap.Int64("rider_id", param.RiderID))
	return nil
}

// UpdateDeliveryStatus 更新配送状态
func (s *riderService) UpdateDeliveryStatus(ctx context.Context, param UpdateDeliveryStatusParam) error {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("更新配送状态参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 校验状态合法性
	validStatus := []string{"待取餐", "配送中", "已完成"}
	if !utils.ContainsString(validStatus, param.DeliveryStatus) {
		return utils.NewParamError("配送状态不合法")
	}

	// 1. 更新配送订单状态
	now := time.Now().Format("2006-01-02 15:04:05")
	if err := s.riderRepo.UpdateDeliveryOrder(ctx, param.OrderID, param.RiderID, param.DeliveryStatus, now); err != nil {
		return err
	}

	// 2. 同步更新订单服务状态
	var orderStatus string
	switch param.DeliveryStatus {
	case "配送中":
		orderStatus = "配送中"
	case "已完成":
		orderStatus = "已完成"
	default:
		orderStatus = param.DeliveryStatus
	}

	updateStatusReq := &orderProto.UpdateOrderStatusRequest{
		OrderId:  param.OrderID,
		Status:   orderStatus,
		Operator: "rider_" + strconv.FormatInt(param.RiderID, 10),
	}
	_, err := client.OrderClient.UpdateOrderStatus(ctx, updateStatusReq)
	if err != nil {
		zap.L().Error("调用订单服务更新状态失败", zap.Int64("order_id", param.OrderID), zap.Error(err))
		return utils.NewSystemError("更新配送状态失败，订单服务异常")
	}

	zap.L().Info("更新配送状态成功", zap.Int64("order_id", param.OrderID), zap.String("status", param.DeliveryStatus))
	return nil
}

// ListPendingOrders 查询待接订单列表
func (s *riderService) ListPendingOrders(ctx context.Context, param ListPendingOrdersParam) (ListOrdersResult, error) {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询待接订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return ListOrdersResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 查询待接订单
	orders, total, err := s.riderRepo.ListPendingOrders(ctx, param.Area, param.Page, param.PageSize)
	if err != nil {
		return ListOrdersResult{}, err
	}

	// 转换结果
	var resultOrders []DeliveryOrderResult
	for _, o := range orders {
		resultOrders = append(resultOrders, DeliveryOrderResult{
			OrderID:        o.OrderID,
			OrderNo:        o.OrderNo,
			RiderID:        o.RiderID,
			RiderName:      o.RiderName,
			MerchantID:     o.MerchantID,
			MerchantName:   o.MerchantName,
			Address:        o.Address,
			TotalAmount:    o.TotalAmount,
			DeliveryStatus: o.DeliveryStatus,
			AcceptTime:     o.AcceptTime,
			PickupTime:     o.PickupTime,
			CompleteTime:   o.CompleteTime,
		})
	}

	result := ListOrdersResult{
		Orders:   resultOrders,
		Total:    int32(total),
		Page:     param.Page,
		PageSize: param.PageSize,
	}

	return result, nil
}

// ListRiderOrders 查询骑手配送订单
func (s *riderService) ListRiderOrders(ctx context.Context, param ListRiderOrdersParam) (ListOrdersResult, error) {
	// 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询骑手订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return ListOrdersResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 查询骑手订单
	orders, total, err := s.riderRepo.ListRiderOrders(ctx, param.RiderID, param.DeliveryStatus, param.Page, param.PageSize)
	if err != nil {
		return ListOrdersResult{}, err
	}

	// 转换结果
	var resultOrders []DeliveryOrderResult
	for _, o := range orders {
		resultOrders = append(resultOrders, DeliveryOrderResult{
			OrderID:        o.OrderID,
			OrderNo:        o.OrderNo,
			RiderID:        o.RiderID,
			RiderName:      o.RiderName,
			MerchantID:     o.MerchantID,
			MerchantName:   o.MerchantName,
			Address:        o.Address,
			TotalAmount:    o.TotalAmount,
			DeliveryStatus: o.DeliveryStatus,
			AcceptTime:     o.AcceptTime,
			PickupTime:     o.PickupTime,
			CompleteTime:   o.CompleteTime,
		})
	}

	result := ListOrdersResult{
		Orders:   resultOrders,
		Total:    int32(total),
		Page:     param.Page,
		PageSize: param.PageSize,
	}

	return result, nil
}
