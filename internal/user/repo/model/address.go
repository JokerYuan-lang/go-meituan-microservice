package model

import (
	"time"

	"gorm.io/gorm"
)

// Address 收货地址表模型
type Address struct {
	AddressID int64          `gorm:"column:address_id;primaryKey;autoIncrement" json:"address_id"`
	UserID    int64          `gorm:"column:user_id;not null;index" json:"user_id"`
	Receiver  string         `gorm:"column:receiver;not null;size:32" json:"receiver"`
	Phone     string         `gorm:"column:phone;not null;size:16" json:"phone"`
	Province  string         `gorm:"column:province;not null;size:16" json:"province"`
	City      string         `gorm:"column:city;not null;size:16" json:"city"`
	District  string         `gorm:"column:district;not null;size:16" json:"district"`
	Detail    string         `gorm:"column:detail;not null;size:255" json:"detail"`
	IsDefault bool           `gorm:"column:is_default;not null;default:false" json:"is_default"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (a *Address) TableName() string {
	return "t_address"
}
