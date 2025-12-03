package service

import (
	"context"
	"strconv"

	userProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/user/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate = validator.New()

type UserService interface {
	Register(ctx context.Context, req *userProto.RegisterRequest) (*userProto.RegisterResponse, error)
	Login(ctx context.Context, req *userProto.LoginRequest) (*userProto.LoginResponse, error)
	GetUserInfo(ctx context.Context, req *userProto.GetUserInfoRequest) (*userProto.GetUserInfoResponse, error)
	UpdateUserInfo(ctx context.Context, req *userProto.UpdateUserInfoRequest) (*userProto.UpdateUserInfoResponse, error)
	AddAddress(ctx context.Context, req *userProto.AddAddressRequest) (*userProto.AddAddressResponse, error)
	ListAddresses(ctx context.Context, req *userProto.ListAddressesRequest) (*userProto.ListAddressesResponse, error)
	UpdateAddress(ctx context.Context, req *userProto.UpdateAddressRequest) (*userProto.UpdateAddressResponse, error)
	DeleteAddress(ctx context.Context, req *userProto.DeleteAddressRequest) (*userProto.DeleteAddressResponse, error)
	SetDefaultAddress(ctx context.Context, req *userProto.SetDefaultAddressRequest) (*userProto.SetDefaultAddressResponse, error)
}

type userService struct {
	userRepo    repo.UserRepo
	addressRepo repo.AddressRepo
}

func NewUserService() UserService {
	return &userService{
		userRepo:    repo.NewUserRepo(),
		addressRepo: repo.NewAddressRepo(),
	}
}

func (u *userService) Register(ctx context.Context, req *userProto.RegisterRequest) (*userProto.RegisterResponse, error) {
	if err := validate.Struct(req); err != nil {
		zap.L().Warn("注册参数校验失败", zap.Any("req", req), zap.Error(err))
		return &userProto.RegisterResponse{
			Code: utils.ErrCodeParam,
			Msg:  "参数错误:" + err.Error(),
		}, err
	}
	//校验手机号是否已经被注册
	existUser, err := u.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return &userProto.RegisterResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, err
	}
	if existUser != nil {
		return &userProto.RegisterResponse{
			Code: utils.ErrCodeBiz,
			Msg:  "手机号已被注册",
		}, nil
	}
	user := &model.User{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
		Role:     "user", //默认为普通用户
	}
	if err = u.userRepo.CreateUser(ctx, user); err != nil {
		return &userProto.RegisterResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	//jwtClaims := &utils.UserClaims{
	//	Username: user.Username,
	//	UserID:   strconv.FormatInt(user.UserID, 10),
	//	Phone:    user.Phone,
	//	Role:     user.Role,
	//}
	//token, err := utils.GenerateToken(jwtClaims)
	//if err != nil {
	//	zap.L().Error("生成注册Token失败", zap.Int64("user_id", user.UserID), zap.Error(err))
	//	return &userProto.RegisterResponse{
	//		Code: int32(err.(*utils.AppError).Code),
	//		Msg:  "注册成功但是生成Token失败",
	//	}, nil
	//}
	zap.L().Info("用户注册成功", zap.Int64("user_id", user.UserID), zap.String("phone", req.Phone))
	return &userProto.RegisterResponse{
		Code:   int32(err.(*utils.AppError).Code),
		Msg:    "注册成功",
		UserId: user.UserID,
	}, nil
}

// Login 用户登录
func (u *userService) Login(ctx context.Context, req *userProto.LoginRequest) (*userProto.LoginResponse, error) {
	if err := validate.Struct(req); err != nil {
		zap.L().Warn("登录参数校验失败", zap.Any("req", req), zap.Error(err))
		return &userProto.LoginResponse{
			Code: utils.ErrCodeParam,
			Msg:  "参数错误" + err.Error(),
		}, nil
	}
	//查询用户
	user, err := u.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return &userProto.LoginResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	if user == nil {
		return &userProto.LoginResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  "手机号或密码出错",
		}, nil
	}
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return &userProto.LoginResponse{
			Code: utils.ErrCodeBiz,
			Msg:  "手机号或密码出错",
		}, nil
	}
	jwtClaims := &utils.UserClaims{
		Username: user.Username,
		UserID:   strconv.FormatInt(user.UserID, 10),
		Phone:    user.Phone,
		Role:     user.Role,
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成登录Token失败", zap.Int64("user_id", user.UserID), zap.Error(err))
		return &userProto.LoginResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  "登录成功，生成Token失败",
		}, nil
	}
	zap.L().Info("用户登录成功", zap.Int64("user_id", user.UserID), zap.String("phone", req.Phone))
	return &userProto.LoginResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "登录成功",
		Token:    token,
		UserId:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (u *userService) GetUserInfo(ctx context.Context, req *userProto.GetUserInfoRequest) (*userProto.GetUserInfoResponse, error) {
	if req.UserId == 0 {
		return &userProto.GetUserInfoResponse{
			Code: utils.ErrCodeParam,
			Msg:  "用户ID不能为空",
		}, nil
	}

	user, err := u.userRepo.GetUserByUserID(ctx, req.UserId)
	if err != nil {
		return &userProto.GetUserInfoResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	if user == nil {
		return &userProto.GetUserInfoResponse{
			Code: utils.ErrCodeBiz,
			Msg:  "用户不存在",
		}, nil
	}
	return &userProto.GetUserInfoResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "查询成功",
		Data: &userProto.UserInfo{
			UserId:    user.UserID,
			Username:  user.Username,
			Phone:     user.Phone,
			Avatar:    user.Avatar,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

func (u *userService) UpdateUserInfo(ctx context.Context, req *userProto.UpdateUserInfoRequest) (*userProto.UpdateUserInfoResponse, error) {
	if req.UserId == 0 {
		return &userProto.UpdateUserInfoResponse{
			Code: utils.ErrCodeParam,
			Msg:  "用户ID不能为空",
		}, nil
	}
	user := &model.User{
		Username: req.Username,
		UserID:   req.UserId,
		Avatar:   req.Avatar,
	}

	err := u.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return &userProto.UpdateUserInfoResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	return &userProto.UpdateUserInfoResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新用户信息成功",
	}, nil
}

func (u *userService) AddAddress(ctx context.Context, req *userProto.AddAddressRequest) (*userProto.AddAddressResponse, error) {
	if req.UserId == 0 {
		return &userProto.AddAddressResponse{
			Code: utils.ErrCodeParam,
			Msg:  "user_id 不能为空",
		}, nil
	}
	address := &model.Address{
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}

	err := u.addressRepo.CreateAddress(ctx, address)
	if err != nil {
		return &userProto.AddAddressResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	return &userProto.AddAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "添加地址成功",
	}, nil
}

func (u *userService) ListAddresses(ctx context.Context, req *userProto.ListAddressesRequest) (*userProto.ListAddressesResponse, error) {
	if req.UserId == 0 {
		return &userProto.ListAddressesResponse{
			Code: utils.ErrCodeParam,
			Msg:  "user_if不能为空",
		}, nil
	}
	addresses, err := u.addressRepo.ListAddressesByUserID(ctx, req.UserId)
	if err != nil {
		return &userProto.ListAddressesResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	addressList := make([]*userProto.Address, 0)
	for _, address := range addresses {
		addressList = append(addressList, &userProto.Address{
			AddressId: address.AddressID,
			UserId:    address.UserID,
			Phone:     address.Phone,
			Province:  address.Province,
			City:      address.City,
			District:  address.District,
			Detail:    address.Detail,
			Receiver:  address.Receiver,
			IsDefault: address.IsDefault,
		})
	}
	return &userProto.ListAddressesResponse{
		Code:      utils.ErrCodeSuccess,
		Msg:       "查询地址列表成功",
		Addresses: addressList,
	}, nil
}

func (u *userService) UpdateAddress(ctx context.Context, req *userProto.UpdateAddressRequest) (*userProto.UpdateAddressResponse, error) {
	if req.UserId == 0 || req.AddressId == 0 {
		return &userProto.UpdateAddressResponse{
			Code: utils.ErrCodeParam,
			Msg:  "用户ID和地址ID不能为空",
		}, nil
	}
	address := &model.Address{
		AddressID: req.AddressId,
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}
	err := u.addressRepo.UpdateAddress(ctx, address)
	if err != nil {
		return &userProto.UpdateAddressResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	return &userProto.UpdateAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新地址成功",
	}, nil
}
func (u *userService) DeleteAddress(ctx context.Context, req *userProto.DeleteAddressRequest) (*userProto.DeleteAddressResponse, error) {
	if req.UserId == 0 || req.AddressId == 0 {
		return &userProto.DeleteAddressResponse{
			Code: utils.ErrCodeParam,
			Msg:  "用户ID和地址ID不能为空",
		}, nil
	}
	err := u.addressRepo.DeleteAddress(ctx, req.AddressId, req.UserId)
	if err != nil {
		return &userProto.DeleteAddressResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	return &userProto.DeleteAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "删除地址成功",
	}, nil
}

func (u *userService) SetDefaultAddress(ctx context.Context, req *userProto.SetDefaultAddressRequest) (*userProto.SetDefaultAddressResponse, error) {
	if req.UserId == 0 || req.AddressId == 0 {
		return &userProto.SetDefaultAddressResponse{
			Code: utils.ErrCodeParam,
			Msg:  "用户ID和地址ID不能为空",
		}, nil
	}
	err := u.addressRepo.UpdateDefaultAddress(ctx, req.AddressId, req.UserId)
	if err != nil {
		return &userProto.SetDefaultAddressResponse{
			Code: int32(err.(*utils.AppError).Code),
			Msg:  err.Error(),
		}, nil
	}
	return &userProto.SetDefaultAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "设置默认地址成功",
	}, nil
}
