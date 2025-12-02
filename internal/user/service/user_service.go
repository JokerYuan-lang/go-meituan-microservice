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

type UserService interface{}

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
	
}

func (u *userService) UpdateUserInfo(ctx context.Context, req *userProto.UpdateUserInfoRequest) (*userProto.UpdateUserInfoResponse, error) {

}

func (u *userService) AddAddress(ctx context.Context, req *userProto.AddAddressRequest) (*userProto.AddAddressResponse, error) {

}

func (u *userService) ListAddresses(ctx context.Context, req *userProto.ListAddressesRequest) (*userProto.ListAddressesResponse, error) {

}
