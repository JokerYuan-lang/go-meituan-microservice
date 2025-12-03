package model

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品表模型
type Product struct {
	ProductID   int64          `gorm:"column:product_id;primaryKey;autoIncrement" json:"product_id"`
	MerchantID  int64          `gorm:"column:merchant_id;not null;index;comment:'商家ID'" json:"merchant_id"`
	Name        string         `gorm:"column:name;not null;size:64;comment:'商品名称'" json:"name"`
	Description string         `gorm:"column:description;size:512;comment:'商品描述'" json:"description"`
	Price       float64        `gorm:"column:price;not null;type:decimal(10,2);comment:'商品价格（元）'" json:"price"`
	Stock       int32          `gorm:"column:stock;not null;default:0;comment:'库存数量'" json:"stock"`
	ImageURL    string         `gorm:"column:image_url;size:255;comment:'商品图片'" json:"image_url"`
	IsSoldOut   bool           `gorm:"column:is_sold_out;not null;default:false;comment:'是否售罄'" json:"is_sold_out"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:'更新时间'" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (p *Product) TableName() string {
	return "t_product"
}

// BeforeUpdate 钩子：更新时自动设置售罄状态
func (p *Product) BeforeUpdate(tx *gorm.DB) error {
	// 库存为0时设为售罄，否则设为未售罄
	if p.Stock <= 0 {
		p.IsSoldOut = true
	} else {
		p.IsSoldOut = false
	}
	return nil
}
