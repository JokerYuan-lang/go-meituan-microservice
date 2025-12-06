package model

import (
	"time"

	"gorm.io/gorm"
)

// DeliveryOrder 配送订单表（关联订单和骑手）
type DeliveryOrder struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID        int64          `gorm:"column:order_id;not null;uniqueIndex;comment:'订单ID'" json:"order_id"`
	OrderNo        string         `gorm:"column:order_no;not null;size:64;comment:'订单编号'" json:"order_no"`
	RiderID        int64          `gorm:"column:rider_id;not null;index;comment:'骑手ID'" json:"rider_id"`
	RiderName      string         `gorm:"column:rider_name;not null;size:64;comment:'骑手姓名'" json:"rider_name"`
	MerchantID     int64          `gorm:"column:merchant_id;not null;comment:'商家ID'" json:"merchant_id"`
	MerchantName   string         `gorm:"column:merchant_name;not null;size:64;comment:'商家名称'" json:"merchant_name"`
	Address        string         `gorm:"column:address;not null;size:255;comment:'配送地址'" json:"address"`
	TotalAmount    float64        `gorm:"column:total_amount;not null;type:decimal(10,2);comment:'订单金额'" json:"total_amount"`
	DeliveryStatus string         `gorm:"column:delivery_status;not null;size:16;default:'待取餐';comment:'配送状态'" json:"delivery_status"`
	AcceptTime     string         `gorm:"column:accept_time;size:32;comment:'接单时间'" json:"accept_time"`
	PickupTime     string         `gorm:"column:pickup_time;size:32;comment:'取餐时间'" json:"pickup_time"`
	CompleteTime   string         `gorm:"column:complete_time;size:32;comment:'完成时间'" json:"complete_time"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:'更新时间'" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (d *DeliveryOrder) TableName() string {
	return "t_delivery_order"
}
