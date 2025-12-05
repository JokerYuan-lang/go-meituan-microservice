package model

import (
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"gorm.io/gorm"
)

type Merchant struct {
	MerchantID    int64          `gorm:"column:merchant_id;primaryKey;autoIncrement" json:"merchant_id"`
	Name          string         `gorm:"column:name;not null;size:64;comment:'商家名称'" json:"name"`
	Phone         string         `gorm:"column:phone;not null;type:varchar(20);uniqueIndex;comment:'商家电话'" json:"phone"`
	Password      string         `gorm:"column:password;not null;size:255;comment:'密码（bcrypt加密）'" json:"-"` // 前端不返回
	Address       string         `gorm:"column:address;not null;size:255;comment:'商家地址'" json:"address"`
	Logo          string         `gorm:"column:logo;size:255;comment:'商家logo'" json:"logo"`
	BusinessHours string         `gorm:"column:business_hours;not null;size:64;comment:'营业时间'" json:"business_hours"`
	Score         float64        `gorm:"column:score;not null;default:5.0;type:decimal(2,1);comment:'商家评分'" json:"score"`
	OrderCount    int32          `gorm:"column:order_count;not null;default:0;comment:'订单数'" json:"order_count"`
	IsOpen        bool           `gorm:"column:is_open;not null;default:true;comment:'是否营业'" json:"is_open"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:'更新时间'" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (m *Merchant) TableName() string {
	return "t_merchant"
}

// BeforeCreate 钩子：密码加密
func (m *Merchant) BeforeCreate(tx *gorm.DB) error {
	// 调用util的bcrypt加密方法（复用用户服务的工具）
	encryptedPwd, err := utils.BcryptHash(m.Password)
	if err != nil {
		return err
	}
	m.Password = encryptedPwd
	return nil
}
