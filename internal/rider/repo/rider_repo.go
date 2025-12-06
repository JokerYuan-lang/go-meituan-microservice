package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/rider/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RiderRepo 骑手数据访问接口
type RiderRepo interface {
	CreateRider(ctx context.Context, rider *model.Rider) error
	GetRiderByPhone(ctx context.Context, phone string) (*model.Rider, error)
	GetRiderByID(ctx context.Context, riderID int64) (*model.Rider, error)
	UpdateRiderStatus(ctx context.Context, riderID int64, status string) error
	UpdateOrderCount(ctx context.Context, riderID int64, num int32) error

	CreateDeliveryOrder(ctx context.Context, order *model.DeliveryOrder) error
	UpdateDeliveryOrder(ctx context.Context, orderID, riderID int64, status, timeStr string) error
	GetDeliveryOrderByOrderID(ctx context.Context, orderID int64) (*model.DeliveryOrder, error)
	ListPendingOrders(ctx context.Context, area string, page, pageSize int32) ([]*model.DeliveryOrder, int64, error)
	ListRiderOrders(ctx context.Context, riderID int64, status string, page, pageSize int32) ([]*model.DeliveryOrder, int64, error)
}

// riderRepo 实现
type riderRepo struct{}

// NewRiderRepo 创建实例
func NewRiderRepo() RiderRepo {
	return &riderRepo{}
}

// CreateRider 骑手注册
func (r *riderRepo) CreateRider(ctx context.Context, rider *model.Rider) error {
	tx := db.Mysql.WithContext(ctx).Create(rider)
	if tx.Error != nil {
		zap.L().Error("创建骑手失败", zap.Any("rider", rider), zap.Error(tx.Error))
		return utils.NewDBError("骑手注册失败：" + tx.Error.Error())
	}
	return nil
}

// GetRiderByPhone 根据手机号查询骑手
func (r *riderRepo) GetRiderByPhone(ctx context.Context, phone string) (*model.Rider, error) {
	var rider model.Rider
	tx := db.Mysql.WithContext(ctx).Where("phone = ?", phone).First(&rider)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.L().Error("查询骑手失败（手机号）", zap.String("phone", phone), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询骑手失败：" + tx.Error.Error())
	}
	return &rider, nil
}

// GetRiderByID 根据ID查询骑手
func (r *riderRepo) GetRiderByID(ctx context.Context, riderID int64) (*model.Rider, error) {
	var rider model.Rider
	tx := db.Mysql.WithContext(ctx).Where("rider_id = ?", riderID).First(&rider)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError("骑手不存在")
		}
		zap.L().Error("查询骑手失败（ID）", zap.Int64("rider_id", riderID), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询骑手失败：" + tx.Error.Error())
	}
	return &rider, nil
}

// UpdateRiderStatus 更新骑手状态
func (r *riderRepo) UpdateRiderStatus(ctx context.Context, riderID int64, status string) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Rider{}).
		Where("rider_id = ?", riderID).
		Update("status", status)
	if tx.Error != nil {
		zap.L().Error("更新骑手状态失败", zap.Int64("rider_id", riderID), zap.String("status", status), zap.Error(tx.Error))
		return utils.NewDBError("更新骑手状态失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("骑手不存在")
	}
	return nil
}

// UpdateOrderCount 更新骑手配送订单数
func (r *riderRepo) UpdateOrderCount(ctx context.Context, riderID int64, num int32) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Rider{}).
		Where("rider_id = ?", riderID).
		Update("order_count", gorm.Expr("order_count + ?", num))
	if tx.Error != nil {
		zap.L().Error("更新骑手订单数失败", zap.Int64("rider_id", riderID), zap.Int32("num", num), zap.Error(tx.Error))
		return utils.NewDBError("更新订单数失败：" + tx.Error.Error())
	}
	return nil
}

// CreateDeliveryOrder 创建配送订单
func (r *riderRepo) CreateDeliveryOrder(ctx context.Context, order *model.DeliveryOrder) error {
	tx := db.Mysql.WithContext(ctx).Create(order)
	if tx.Error != nil {
		zap.L().Error("创建配送订单失败", zap.Any("order", order), zap.Error(tx.Error))
		return utils.NewDBError("创建配送订单失败：" + tx.Error.Error())
	}
	return nil
}

// UpdateDeliveryOrder 更新配送订单状态
func (r *riderRepo) UpdateDeliveryOrder(ctx context.Context, orderID, riderID int64, status, timeStr string) error {
	updateData := map[string]interface{}{
		"delivery_status": status,
	}
	switch status {
	case "待取餐":
		updateData["accept_time"] = timeStr
		updateData["rider_id"] = riderID
		updateData["rider_name"] = "" // 后续补充骑手姓名
	case "配送中":
		updateData["pickup_time"] = timeStr
	case "已完成":
		updateData["complete_time"] = timeStr
	}

	tx := db.Mysql.WithContext(ctx).Model(&model.DeliveryOrder{}).
		Where("order_id = ?", orderID).
		Updates(updateData)
	if tx.Error != nil {
		zap.L().Error("更新配送订单状态失败", zap.Int64("order_id", orderID), zap.String("status", status), zap.Error(tx.Error))
		return utils.NewDBError("更新配送状态失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("配送订单不存在")
	}
	return nil
}

// GetDeliveryOrderByOrderID 根据订单ID查询配送订单
func (r *riderRepo) GetDeliveryOrderByOrderID(ctx context.Context, orderID int64) (*model.DeliveryOrder, error) {
	var order model.DeliveryOrder
	tx := db.Mysql.WithContext(ctx).Where("order_id = ?", orderID).First(&order)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError("配送订单不存在")
		}
		zap.L().Error("查询配送订单失败", zap.Int64("order_id", orderID), zap.Error(tx.Error))
		return nil, utils.NewDBError("查询配送订单失败：" + tx.Error.Error())
	}
	return &order, nil
}

// ListPendingOrders 查询待接订单列表（未分配骑手的订单）
func (r *riderRepo) ListPendingOrders(ctx context.Context, area string, page, pageSize int32) ([]*model.DeliveryOrder, int64, error) {
	var (
		orders []*model.DeliveryOrder
		total  int64
	)

	// 构建查询条件：未分配骑手（rider_id=0）
	query := db.Mysql.WithContext(ctx).Model(&model.DeliveryOrder{}).Where("rider_id = 0")
	if area != "" {
		query = query.Where("address LIKE ?", "%"+area+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("统计待接订单总数失败", zap.String("area", area), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		zap.L().Error("查询待接订单列表失败", zap.String("area", area), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	return orders, total, nil
}

// ListRiderOrders 查询骑手配送订单
func (r *riderRepo) ListRiderOrders(ctx context.Context, riderID int64, status string, page, pageSize int32) ([]*model.DeliveryOrder, int64, error) {
	var (
		orders []*model.DeliveryOrder
		total  int64
	)

	// 构建查询条件
	query := db.Mysql.WithContext(ctx).Model(&model.DeliveryOrder{}).Where("rider_id = ?", riderID)
	if status != "" {
		query = query.Where("delivery_status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("统计骑手订单总数失败", zap.Int64("rider_id", riderID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		zap.L().Error("查询骑手订单列表失败", zap.Int64("rider_id", riderID), zap.Error(err))
		return nil, 0, utils.NewDBError("查询订单失败：" + err.Error())
	}

	return orders, total, nil
}
