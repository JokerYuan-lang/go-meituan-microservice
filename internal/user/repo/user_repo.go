package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	GetUserByUserID(ctx context.Context, userID int64) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
}

type userRepo struct{}

func NewUserRepo() UserRepo {
	return &userRepo{}
}

func (u *userRepo) CreateUser(ctx context.Context, user *model.User) error {
	if err := db.Mysql.WithContext(ctx).Create(&user).Error; err != nil {
		zap.L().Error("创建用户失败", zap.Error(err), zap.Any("user", user))
		return err
	}
	return nil
}

func (u *userRepo) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	if err := db.Mysql.WithContext(ctx).First(&user, "phone = ?", phone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Warn("用户不存在", zap.String("phone", phone))
			return nil, nil
		}
		zap.L().Error("根据手机号查找用户失败", zap.Error(err), zap.String("phone", phone))
		return nil, err
	}
	return &user, nil
}

func (u *userRepo) GetUserByUserID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := db.Mysql.WithContext(ctx).First(&user, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Warn("用户不存在", zap.Int64("user_id", userID))
			return nil, nil
		}
		zap.L().Error("根据用户ID查找用户失败", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}
	return &user, nil
}

func (u *userRepo) UpdateUser(ctx context.Context, user *model.User) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.User{}).Where("user_id = ?", user.UserID).Updates(user)
	if err := tx.Error; err != nil {
		zap.L().Error("更新用户失败", zap.Any("user", user), zap.Error(tx.Error))
		return utils.NewDBError("更新用户失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		zap.L().Warn("更新用户无数据变化", zap.Int64("user_id", user.UserID))
		return utils.NewBizError("无数据更新")
	}
	return nil
}
