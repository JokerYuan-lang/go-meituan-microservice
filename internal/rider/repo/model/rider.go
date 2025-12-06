package model

import (
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"gorm.io/gorm"
)

// Rider 骑手表模型
type Rider struct {
	RiderID    int64          `gorm:"column:rider_id;primaryKey;autoIncrement" json:"rider_id"`
	Name       string         `gorm:"column:name;not null;size:64;comment:'骑手姓名'" json:"name"`
	Phone      string         `gorm:"column:phone;not null;uniqueIndex;comment:'骑手电话'" json:"phone"`
	Password   string         `gorm:"column:password;not null;size:255;comment:'密码（bcrypt加密）'" json:"-"`
	Avatar     string         `gorm:"column:avatar;size:255;comment:'骑手头像'" json:"avatar"`
	Score      float64        `gorm:"column:score;not null;default:5.0;type:decimal(2,1);comment:'骑手评分'" json:"score"`
	OrderCount int32          `gorm:"column:order_count;not null;default:0;comment:'配送订单数'" json:"order_count"`
	Status     string         `gorm:"column:status;not null;size:16;default:'在线';comment:'骑手状态'" json:"status"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:'更新时间'" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 表名
func (r *Rider) TableName() string {
	return "t_rider"
}

// BeforeCreate 钩子：密码加密
func (r *Rider) BeforeCreate(tx *gorm.DB) error {
	encryptedPwd, err := utils.BcryptHash(r.Password)
	if err != nil {
		return err
	}
	r.Password = encryptedPwd
	return nil
}
