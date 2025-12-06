package service

import (
	"context"
	"strconv"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/client"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/repo/model"
	orderProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/order/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 入参结构体（领域层）
type MerchantRegisterParam struct {
	Name          string `validate:"required,min=2,max=64"`
	Phone         string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Password      string `validate:"required,min=6,max=20"`
	Address       string `validate:"required,min=5,max=255"`
	Logo          string `validate:"required,url"`
	BusinessHours string `validate:"required,min=5"`
}

type MerchantLoginParam struct {
	Phone    string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Password string `validate:"required,min=6,max=20"`
}

type UpdateMerchantInfoParam struct {
	MerchantID    int64  `validate:"required,gt=0"`
	Name          string `validate:"required,min=2,max=64"`
	Phone         string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Address       string `validate:"required,min=5,max=255"`
	Logo          string `validate:"required,url"`
	BusinessHours string `validate:"required,min=5"`
	IsOpen        bool   `validate:"required"`
}

type AcceptOrderParam struct {
	OrderID    int64 `validate:"required,gt=0"`
	MerchantID int64 `validate:"required,gt=0"`
}

type RejectOrderParam struct {
	OrderID    int64  `validate:"required,gt=0"`
	MerchantID int64  `validate:"required,gt=0"`
	Reason     string `validate:"required,min=2,max=128"`
}

type ListMerchantOrdersParam struct {
	MerchantID int64  `validate:"required,gt=0"`
	Status     string `validate:"omitempty"`
	Page       int32  `validate:"required,gte=1"`
	PageSize   int32  `validate:"required,gte=10,lte=100"`
}

// 响应结构体（领域层）
type MerchantLoginResult struct {
	MerchantID int64  `json:"merchant_id"`
	Name       string `json:"name"`
	Token      string `json:"token"`
}

type MerchantInfoResult struct {
	MerchantID    int64   `json:"merchant_id"`
	Name          string  `json:"name"`
	Phone         string  `json:"phone"`
	Address       string  `json:"address"`
	Logo          string  `json:"logo"`
	BusinessHours string  `json:"business_hours"`
	Score         float64 `json:"score"`
	OrderCount    int32   `json:"order_count"`
	IsOpen        bool    `json:"is_open"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type MerchantOrderResult struct {
	OrderID            int64   `json:"order_id"`
	UserID             int64   `json:"user_id"`
	UserName           string  `json:"user_name"`
	UserPhone          string  `json:"user_phone"`
	TotalAmount        float64 `json:"total_amount"`
	Status             string  `json:"status"`
	CreateTime         string  `json:"create_time"`
	ExpectDeliveryTime string  `json:"expect_delivery_time"`
}

type ListMerchantOrdersResult struct {
	Orders   []MerchantOrderResult `json:"orders"`
	Total    int32                 `json:"total"`
	Page     int32                 `json:"page"`
	PageSize int32                 `json:"page_size"`
}

// MerchantService 商家业务逻辑接口
type MerchantService interface {
	MerchantRegister(ctx context.Context, param MerchantRegisterParam) (int64, string, error) // 返回商家ID、Token、错误
	MerchantLogin(ctx context.Context, param MerchantLoginParam) (MerchantLoginResult, error)
	GetMerchantInfo(ctx context.Context, merchantID int64) (MerchantInfoResult, error)
	UpdateMerchantInfo(ctx context.Context, param UpdateMerchantInfoParam) error
	AcceptOrder(ctx context.Context, param AcceptOrderParam) error // 接单
	RejectOrder(ctx context.Context, param RejectOrderParam) error // 拒单
	ListMerchantOrders(ctx context.Context, param ListMerchantOrdersParam) (ListMerchantOrdersResult, error)
}

// merchantService 实现
type merchantService struct {
	merchantRepo repo.MerchantRepo
	validate     *validator.Validate
}

// NewMerchantService 创建实例
func NewMerchantService(merchantRepo repo.MerchantRepo) MerchantService {
	return &merchantService{
		merchantRepo: merchantRepo,
		validate:     validator.New(),
	}
}

// MerchantRegister 商家入驻（注册）
func (s *merchantService) MerchantRegister(ctx context.Context, param MerchantRegisterParam) (int64, string, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("商家入驻参数校验失败", zap.Any("param", param), zap.Error(err))
		return 0, "", utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 校验手机号是否已注册
	existMerchant, err := s.merchantRepo.GetMerchantByPhone(ctx, param.Phone)
	if err != nil {
		return 0, "", err
	}
	if existMerchant != nil {
		return 0, "", utils.NewBizError("手机号已注册")
	}

	// 3. 转换为模型
	merchant := &model.Merchant{
		Name:          param.Name,
		Phone:         param.Phone,
		Password:      param.Password, // BeforeCreate钩子自动加密
		Address:       param.Address,
		Logo:          param.Logo,
		BusinessHours: param.BusinessHours,
		IsOpen:        true, // 默认营业
	}

	// 4. 调用Repo创建商家
	if err = s.merchantRepo.CreateMerchant(ctx, merchant); err != nil {
		return 0, "", err
	}

	// 5. 生成JWT Token（商家角色）
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(merchant.MerchantID, 10),
		Username: merchant.Name,
		Phone:    merchant.Phone,
		Role:     "merchant", // 商家角色
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成商家Token失败", zap.Int64("merchant_id", merchant.MerchantID), zap.Error(err))
		return merchant.MerchantID, "", utils.NewSystemError("入驻成功，但生成Token失败")
	}

	zap.L().Info("商家入驻成功", zap.Int64("merchant_id", merchant.MerchantID), zap.String("phone", param.Phone))
	return merchant.MerchantID, token, nil
}

// MerchantLogin 商家登录
func (s *merchantService) MerchantLogin(ctx context.Context, param MerchantLoginParam) (MerchantLoginResult, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("商家登录参数校验失败", zap.Any("param", param), zap.Error(err))
		return MerchantLoginResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 查询商家
	merchant, err := s.merchantRepo.GetMerchantByPhone(ctx, param.Phone)
	if err != nil {
		return MerchantLoginResult{}, err
	}
	if merchant == nil {
		return MerchantLoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 3. 验证密码
	if !utils.CheckPasswordHash(param.Password, merchant.Password) {
		return MerchantLoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 4. 生成Token
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(merchant.MerchantID, 10),
		Username: merchant.Name,
		Phone:    merchant.Phone,
		Role:     "merchant",
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成商家登录Token失败", zap.Int64("merchant_id", merchant.MerchantID), zap.Error(err))
		return MerchantLoginResult{}, utils.NewSystemError("登录失败，生成Token失败")
	}

	// 5. 组装结果
	result := MerchantLoginResult{
		MerchantID: merchant.MerchantID,
		Name:       merchant.Name,
		Token:      token,
	}

	zap.L().Info("商家登录成功", zap.Int64("merchant_id", merchant.MerchantID), zap.String("phone", param.Phone))
	return result, nil
}

// GetMerchantInfo 获取商家信息
func (s *merchantService) GetMerchantInfo(ctx context.Context, merchantID int64) (MerchantInfoResult, error) {
	// 1. 参数校验
	if merchantID <= 0 {
		return MerchantInfoResult{}, utils.NewParamError("商家ID不能为空且大于0")
	}

	// 2. 查询商家
	merchant, err := s.merchantRepo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		return MerchantInfoResult{}, err
	}

	// 3. 组装结果
	result := MerchantInfoResult{
		MerchantID:    merchant.MerchantID,
		Name:          merchant.Name,
		Phone:         merchant.Phone,
		Address:       merchant.Address,
		Logo:          merchant.Logo,
		BusinessHours: merchant.BusinessHours,
		Score:         merchant.Score,
		OrderCount:    merchant.OrderCount,
		IsOpen:        merchant.IsOpen,
		CreatedAt:     merchant.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     merchant.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// UpdateMerchantInfo 更新商家信息
func (s *merchantService) UpdateMerchantInfo(ctx context.Context, param UpdateMerchantInfoParam) error {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("更新商家信息参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 转换为模型
	merchant := &model.Merchant{
		MerchantID:    param.MerchantID,
		Name:          param.Name,
		Phone:         param.Phone,
		Address:       param.Address,
		Logo:          param.Logo,
		BusinessHours: param.BusinessHours,
		IsOpen:        param.IsOpen,
	}

	// 3. 调用Repo更新
	return s.merchantRepo.UpdateMerchant(ctx, merchant)
}

// AcceptOrder 商家接单（核心逻辑：后续对接订单服务更新订单状态）
func (s *merchantService) AcceptOrder(ctx context.Context, param AcceptOrderParam) error {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("商家接单参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 校验商家是否存在且营业
	merchant, err := s.merchantRepo.GetMerchantByID(ctx, param.MerchantID)
	if err != nil {
		return err
	}
	if !merchant.IsOpen {
		return utils.NewBizError("商家已歇业，无法接单")
	}

	updateStatusReq := &orderProto.UpdateOrderStatusRequest{
		OrderId:  param.OrderID,
		Status:   "已接单",
		Operator: "merchant_" + strconv.FormatInt(param.MerchantID, 10),
	}

	_, err = client.OrderClient.UpdateOrderStatus(ctx, updateStatusReq)
	if err != nil {
		zap.L().Error("调用订单服务更新状态失败", zap.Any("param", param), zap.Error(err))
		return utils.NewSystemError("接单失败,订单服务异常")
	}

	// 4. 更新商家订单数+1
	if err = s.merchantRepo.UpdateOrderCount(ctx, param.MerchantID, 1); err != nil {
		zap.L().Warn("更新商家订单数失败", zap.Int64("merchant_id", param.MerchantID), zap.Error(err))
		// 不影响接单逻辑，仅日志警告
	}
	zap.L().Info("商家接单成功", zap.Int64("order_id", param.OrderID), zap.Int64("merchant_id", param.MerchantID))
	return nil
}

// RejectOrder 商家拒单（核心逻辑：后续对接订单服务更新订单状态）
func (s *merchantService) RejectOrder(ctx context.Context, param RejectOrderParam) error {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("商家拒单参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 校验商家是否存在
	_, err := s.merchantRepo.GetMerchantByID(ctx, param.MerchantID)
	if err != nil {
		return err
	}

	updateStatusReq := &orderProto.UpdateOrderStatusRequest{
		OrderId:  param.OrderID,
		Status:   "已拒单",
		Operator: "merchant_" + strconv.FormatInt(param.MerchantID, 10),
		Remark:   param.Reason,
	}
	_, err = client.OrderClient.UpdateOrderStatus(ctx, updateStatusReq)
	if err != nil {
		zap.L().Error("调用订单服务更新状态失败", zap.Any("param", param), zap.Error(err))
		return utils.NewSystemError("拒单失败，订单服务异常")
	}

	zap.L().Info("商家拒单", zap.Int64("order_id", param.OrderID), zap.Int64("merchant_id", param.MerchantID), zap.String("reason", param.Reason))

	return nil
}

// ListMerchantOrders 查询商家订单列表（TODO：后续对接订单服务获取真实订单数据）
func (s *merchantService) ListMerchantOrders(ctx context.Context, param ListMerchantOrdersParam) (ListMerchantOrdersResult, error) {
	// 1. 参数校验
	if err := s.validate.Struct(param); err != nil {
		zap.L().Warn("查询商家订单参数校验失败", zap.Any("param", param), zap.Error(err))
		return ListMerchantOrdersResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用订单服务获取订单列表
	listReq := &orderProto.ListMerchantOrdersRequest{
		MerchantId: param.MerchantID,
		Status:     param.Status,
		Page:       param.Page,
		PageSize:   param.PageSize,
	}
	listResp, err := client.OrderClient.ListMerchantOrders(ctx, listReq)
	if err != nil {
		zap.L().Error("调用订单服务查询订单失败", zap.Int64("merchant_id", param.MerchantID), zap.Error(err))
		return ListMerchantOrdersResult{}, utils.NewSystemError("查询订单失败，订单服务异常")
	}
	var orders []MerchantOrderResult
	for _, o := range listResp.Orders {
		orders = append(orders, MerchantOrderResult{
			OrderID:            o.OrderId,
			UserID:             o.UserId,
			UserName:           o.UserName,
			UserPhone:          o.UserPhone,
			TotalAmount:        float64(o.TotalAmount),
			Status:             o.Status,
			CreateTime:         o.CreateTime,
			ExpectDeliveryTime: o.ExpectDeliveryTime,
		})
	}

	result := ListMerchantOrdersResult{
		Orders:   orders,
		Total:    listResp.Total,
		Page:     listResp.Page,
		PageSize: listResp.PageSize,
	}

	return result, nil
}
