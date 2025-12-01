package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表模型
type User struct {
	UserID    int64          `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Username  string         `gorm:"column:username;not null;size:32" json:"username"`
	Password  string         `gorm:"column:password;not null;size:64" json:"-"` // 密码不返回给前端
	Phone     string         `gorm:"column:phone;not null;size:16;uniqueIndex" json:"phone"`
	Avatar    string         `gorm:"column:avatar;size:255;default:'https://ts1.tc.mm.bing.net/th/id/R-C.fba5865e44a8ace919e7308f9bb85297?rik=zFyDrC%2b8%2fdJ0pg&riu=http%3a%2f%2fwww.kzmwh.com%2fuploads%2fimg%2f20201015%2f1602772107267786.png&ehk=wBLs3JBEm0KiSaPJnXlcEHlKJBBlQuYCrahADuZbjjo%3d&risl=&pid=ImgRaw&r=0'" json:"avatar"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"` // 软删除字段
}

// TableName 指定表名
func (u *User) TableName() string {
	return "t_user"
}
