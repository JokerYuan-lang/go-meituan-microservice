package repo

import (
	"context"
	"errors"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AddressRepo interface {
	CreateAddress(ctx context.Context, addr *model.Address) error
	ListAddressesByUserID(ctx context.Context, userID int64) ([]*model.Address, error)
	UpdateDefaultAddress(ctx context.Context, userID int64, addressID int64) error
}

type addressRepo struct {
}

func NewAddressRepo() AddressRepo {
	return &addressRepo{}
}

func (a *addressRepo) CreateAddress(ctx context.Context, addr *model.Address) error {
	if err := db.Mysql.WithContext(ctx).Create(&addr).Error; err != nil {
		zap.L().Fatal("创建地址失败", zap.Error(err))
		return err
	}
	zap.L().Info("创建地址成功", zap.Any("addr", addr))
	return nil
}

func (a *addressRepo) ListAddressesByUserID(ctx context.Context, userID int64) ([]*model.Address, error) {
	addresses := make([]*model.Address, 0)
	if err := db.Mysql.WithContext(ctx).Model(&model.Address{}).Find(&addresses, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Fatal("用户ID 不存在", zap.Int64("user_id", userID))
			return nil, err
		}
		zap.L().Fatal("查找地址列表失败", zap.Error(err))
		return nil, err
	}
	return addresses, nil
}

// UpdateDefaultAddress 更新默认地址
func (a *addressRepo) UpdateDefaultAddress(ctx context.Context, userID int64, addressID int64) error {
	tx := db.Mysql.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&model.Address{}).Where("user_id = ?", userID).Update("id_default", false).Error; err != nil {
		zap.L().Error("更新默认地址失败（重置）", zap.Error(err), zap.Int64("user_id", userID))
		tx.Rollback()
		return err
	}

	if err := tx.Model(&model.Address{}).Where("user_id = ? AND address_id = ?", userID, addressID).Update("default", true).Error; err != nil {
		zap.L().Error("更新默认地址失败（设置）", zap.Error(err), zap.Int64("address_id", addressID), zap.Int64("user_id", userID))
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
