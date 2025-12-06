package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order 订单主表
type Order struct {
	OrderID            int64          `gorm:"column:order_id;primaryKey;autoIncrement" json:"order_id"`
	OrderNo            string         `gorm:"column:order_no;not null;uniqueIndex;size:64;comment:'订单编号'" json:"order_no"`
	UserID             int64          `gorm:"column:user_id;not null;index;comment:'用户ID'" json:"user_id"`
	UserName           string         `gorm:"column:user_name;not null;size:64;comment:'用户名'" json:"user_name"`
	UserPhone          string         `gorm:"column:user_phone;not null;size:11;comment:'用户电话'" json:"user_phone"`
	MerchantID         int64          `gorm:"column:merchant_id;not null;index;comment:'商家ID'" json:"merchant_id"`
	MerchantName       string         `gorm:"column:merchant_name;not null;size:64;comment:'商家名称'" json:"merchant_name"`
	TotalAmount        float64        `gorm:"column:total_amount;not null;type:decimal(10,2);comment:'订单总金额'" json:"total_amount"`
	Status             string         `gorm:"column:status;not null;size:16;default:'待接单';comment:'订单状态'" json:"status"`
	Address            string         `gorm:"column:address;not null;size:255;comment:'收货地址'" json:"address"`
	ExpectDeliveryTime string         `gorm:"column:expect_delivery_time;size:32;comment:'预计送达时间'" json:"expect_delivery_time"`
	Remark             string         `gorm:"column:remark;size:255;comment:'备注'" json:"remark"`
	CreateTime         time.Time      `gorm:"column:create_time;autoCreateTime;comment:'创建时间'" json:"create_time"`
	UpdateTime         time.Time      `gorm:"column:update_time;autoUpdateTime;comment:'更新时间'" json:"update_time"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (o *Order) TableName() string {
	return "t_order"
}

// BeforeCreate 钩子：生成唯一订单编号
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	// 生成规则：YYYYMMDD + uuid
	now := time.Now()
	dateStr := now.Format("20060102")
	u := uuid.New()
	randomStr := u.String() // 复用工具类生成8位随机数
	o.OrderNo = dateStr + randomStr
	return nil
}
