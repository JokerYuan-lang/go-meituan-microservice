package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderItem 订单项表（商品）
type OrderItem struct {
	ItemID      int64          `gorm:"column:item_id;primaryKey;autoIncrement" json:"item_id"`
	OrderID     int64          `gorm:"column:order_id;not null;index;comment:'订单ID'" json:"order_id"`
	ProductID   int64          `gorm:"column:product_id;not null;comment:'商品ID'" json:"product_id"`
	ProductName string         `gorm:"column:product_name;not null;size:64;comment:'商品名称'" json:"product_name"`
	Price       float64        `gorm:"column:price;not null;type:decimal(10,2);comment:'商品单价'" json:"price"`
	Quantity    int32          `gorm:"column:quantity;not null;default:1;comment:'购买数量'" json:"quantity"`
	TotalPrice  float64        `gorm:"column:total_price;not null;type:decimal(10,2);comment:'商品总价'" json:"total_price"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (oi *OrderItem) TableName() string {
	return "t_order_item"
}
