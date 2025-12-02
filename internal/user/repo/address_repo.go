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

type AddressRepo interface {
	CreateAddress(ctx context.Context, addr *model.Address) error
	GetAddressByID(ctx context.Context, addressID int64) (*model.Address, error)
	UpdateAddress(ctx context.Context, addr *model.Address) error
	DeleteAddress(ctx context.Context, addressID, userID int64) error
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

func (a *addressRepo) GetAddressByID(ctx context.Context, addressID int64) (*model.Address, error) {
	var addr *model.Address
	tx := db.Mysql.WithContext(ctx).Where("address_id = ?", addressID).First(addr)
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Warn("地址不存在", zap.Int64("address_id", addressID))
			return nil, nil
		}
		zap.L().Error("查询地址失败", zap.Int64("address_id", addressID), zap.Error(err))
		return nil, utils.NewDBError("查询地址失败" + tx.Error.Error())
	}
	return addr, nil
}

func (a *addressRepo) UpdateAddress(ctx context.Context, addr *model.Address) error {
	tx := db.Mysql.WithContext(ctx).Model(&model.Address{}).Where("address_id = ? AND user_id = ?", addr.AddressID, addr.UserID).Updates(addr)
	if err := tx.Error; err != nil {
		zap.L().Error("更新地址失败", zap.Any("address", addr), zap.Error(tx.Error))
		return utils.NewDBError("更新地址失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewBizError("地址不存在或无数据更新")
	}
	return nil
}

func (a *addressRepo) DeleteAddress(ctx context.Context, addressID, userID int64) error {
	tx := db.Mysql.WithContext(ctx).Where("address_id = ? AND user_id = ?", addressID, userID).Delete(&model.Address{})
	if err := tx.Error; err != nil {
		zap.L().Error("删除地址失败", zap.Any("address_id", addressID), zap.Error(tx.Error))
		return utils.NewDBError("删除地址失败：" + tx.Error.Error())
	}
	if tx.RowsAffected == 0 {
		return utils.NewDBError("地址不存在")
	}
	return nil
}
