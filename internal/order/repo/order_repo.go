package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/order/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderRepo 订单数据访问接口
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *model.Order, items []*model.OrderItem) error // 事务创建订单+订单项
	UpdateOrderStatus(ctx context.Context, orderID int64, status, remark string) error
	ListUserOrders(ctx context.Context, userID int64, status string, page, pageSize int32) ([]*model.Order, int64, error)
	ListMerchantOrders(ctx context.Context, merchantID int64, status string, page, pageSize int32) ([]*model.Order, int64, error)
	GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error)
	CancelOrder(ctx context.Context, orderID, userID int64, reason string) error
	GetOrderItems(ctx context.Context, orderID int64) ([]*model.OrderItem, error) // 查询订单项
}

// orderRepo 实现
type orderRepo struct{}

// NewOrderRepo 创建实例
func NewOrderRepo() OrderRepo {
	return &orderRepo{}
}

// CreateOrder 事务创建订单+订单项
func (r *orderRepo) CreateOrder(ctx context.Context, order *model.Order, items []*model.OrderItem) error {
	tx := db.Mysql.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建订单主表
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		zap.L().Error("创建订单主表失败", zap.Any("order", order), zap.Error(err))
		return utils.NewDBError("创建订单失败：" + err.Error())
	}

	// 2. 批量创建订单项（关联订单ID）
	for _, item := range items {
		item.OrderID = order.OrderID
	}
	if err := tx.CreateInBatches(items, len(items)).Error; err != nil {
		tx.Rollback()
		zap.L().Error("创建订单项失败", zap.Any("items", items), zap.Error(err))
		return utils.NewDBError("创建订单失败：" + err.Error())
	}

	// 3. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.NewDBError("创建订单失败：" + err.Error())
	}

	return nil
}

// UpdateOrderStatus 更新订单状态
func (r *orderRepo) UpdateOrderStatus(ctx context.Context, orderID int64, status, remark string) error {
	updateData := map[string]interface{}{
		"status": status,
	}
	if remark != "" {
		updateData["remark"] = remark
	}

	tx := db.Mysql.WithContext(ctx).Model(&model.Order{}).
		Where("order_id = ?", orderID).
		Updates(updateData)
	if tx.Error != nil {
		zap.L().Error("更新订单状态失败", zap.Int64("order_id", orderID), zap.String("status", status), zap.Error(tx.Error))
		return utils.NewDBError("更新订单状态失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("订单不存在")
	}
	return nil
}

// ListUserOrders 查询用户订单列表
func (r *orderRepo) ListUserOrders(ctx context.Context, userID int64, status string, page, pageSize int32) ([]*model.Order, int64, error) {
	var (
		orders []*model.Order
		total  int64
	)

	// 构建查询条件
	query := db.Mysql.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("统计用户订单总数失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).
		Order("create_time DESC").Find(&orders).Error; err != nil {
		zap.L().Error("查询用户订单列表失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	return orders, total, nil
}

// ListMerchantOrders 查询商家订单列表
func (r *orderRepo) ListMerchantOrders(ctx context.Context, merchantID int64, status string, page, pageSize int32) ([]*model.Order, int64, error) {
	var (
		orders []*model.Order
		total  int64
	)

	// 构建查询条件
	query := db.Mysql.WithContext(ctx).Model(&model.Order{}).Where("merchant_id = ?", merchantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("统计商家订单总数失败", zap.Int64("merchant_id", merchantID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).
		Order("create_time DESC").Find(&orders).Error; err != nil {
		zap.L().Error("查询商家订单列表失败", zap.Int64("merchant_id", merchantID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	return orders, total, nil
}

// GetOrderByID 查询订单详情
func (r *orderRepo) GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error) {
	var order model.Order
	tx := db.Mysql.WithContext(ctx).Where("order_id = ?", orderID).First(&order)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError("订单不存在")
		}
		zap.L().Error("查询订单详情失败", zap.Int64("order_id", orderID), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询订单失败：" + tx.Error.Error())
	}
	return &order, nil
}

// CancelOrder 取消订单（更新状态+备注）
func (r *orderRepo) CancelOrder(ctx context.Context, orderID, userID int64, reason string) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Order{}).
		Where("order_id = ? AND user_id = ?", orderID, userID).
		Updates(map[string]interface{}{
			"status": "已取消",
			"remark": reason,
		})
	if tx.Error != nil {
		zap.L().Error("取消订单失败", zap.Int64("order_id", orderID), zap.Int64("user_id", userID), zap.Error(tx.Error))
		return utils.NewDBError("取消订单失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("订单不存在或无权限取消")
	}
	return nil
}

// GetOrderItems 查询订单项
func (r *orderRepo) GetOrderItems(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	if err := db.Mysql.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		zap.L().Error("查询订单项失败", zap.Int64("order_id", orderID), zap.Error(err))
		return nil, utils.NewDBError("查询订单项失败：" + err.Error())
	}
	return items, nil
}
