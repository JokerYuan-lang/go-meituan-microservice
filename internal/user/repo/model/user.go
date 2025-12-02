package model

import (
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"gorm.io/gorm"
)

// User 用户表模型
type User struct {
	UserID    int64          `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Username  string         `gorm:"column:username;not null;size:32;comment:'用户名'" json:"username"`
	Password  string         `gorm:"column:password;not null;size:128;comment:'bcrypt加密后的密码'" json:"-"` // 前端不返回密码
	Phone     string         `gorm:"column:phone;not null;size:16;uniqueIndex;comment:'手机号'" json:"phone"`
	Avatar    string         `gorm:"column:avatar;size:255;default:'https://picsum.photos/200';comment:'头像'" json:"avatar"`
	Role      string         `gorm:"column:role;not null;size:16;default:'user';comment:'角色：user/merchant/rider'" json:"role"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:'更新时间'" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:'软删除时间'" json:"-"`
}

// TableName 指定表名
func (u *User) TableName() string {
	return "t_user"
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.Changed("password") {
		encryptPwd, err := utils.BcryptHash(u.Password)
		if err != nil {
			return err
		}
		u.Password = encryptPwd
	}
	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	if u.Avatar == "" {
		u.Avatar = "https://ts1.tc.mm.bing.net/th/id/R-C.54d0da7ab73d8e264c031ef1ce1421de?rik=ABKcbe849Ve92Q&riu=http%3a%2f%2fn.sinaimg.cn%2fsinacn01%2f593%2fw641h752%2f20181122%2f8f8b-hnyuqhi7197762.png&ehk=RXjmmVz4Cozi5iRM2f8h%2fc4SPjwb9Pjh9Dc5Cutxlos%3d&risl=&pid=ImgRaw&r=0"
	}
	return nil
}
