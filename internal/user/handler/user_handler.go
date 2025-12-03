package handler

import (
	"context"
	"errors"

	userProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/user/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/service"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
)

// UserHandler gRPC接口实现（仅做转换和调用service，不写业务逻辑）
type UserHandler struct {
	userProto.UnimplementedUserServiceServer                     // 必须嵌入，兼容proto3
	userService                              service.UserService // 依赖service接口，不依赖具体实现
}

// NewUserHandler 创建UserHandler实例（依赖注入service）
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register gRPC注册接口（proto请求→service入参→service调用→proto响应）
func (h *UserHandler) Register(ctx context.Context, req *userProto.RegisterRequest) (*userProto.RegisterResponse, error) {
	// 1. proto请求 → service入参（RegisterParam）转换
	param := service.RegisterParam{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
	}

	// 2. 调用service层方法
	userID, token, err := h.userService.Register(ctx, param)

	// 3. 错误处理（转换为proto响应的code和msg）
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			// 未知错误
			zap.L().Error("注册接口未知错误", zap.Error(err))
			return &userProto.RegisterResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		// 已知错误（参数错误、业务错误等）
		return &userProto.RegisterResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. service返回结果 → proto响应转换
	return &userProto.RegisterResponse{
		Code:   utils.ErrCodeSuccess,
		Msg:    "注册成功",
		UserId: userID,
		Token:  token,
	}, nil
}

// Login gRPC登录接口（proto请求→service入参→service调用→proto响应）
func (h *UserHandler) Login(ctx context.Context, req *userProto.LoginRequest) (*userProto.LoginResponse, error) {
	// 1. proto请求 → service入参（LoginParam）转换
	param := service.LoginParam{
		Phone:    req.Phone,
		Password: req.Password,
	}

	// 2. 调用service层方法
	result, err := h.userService.Login(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("登录接口未知错误", zap.Error(err))
			return &userProto.LoginResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.LoginResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. service返回结果 → proto响应转换
	return &userProto.LoginResponse{
		Code:     utils.ErrCodeSuccess,
		Msg:      "登录成功",
		Token:    result.Token,
		UserId:   result.UserID,
		Username: result.Username,
		Role:     result.Role,
	}, nil
}

// GetUserInfo gRPC获取用户信息接口
func (h *UserHandler) GetUserInfo(ctx context.Context, req *userProto.GetUserInfoRequest) (*userProto.GetUserInfoResponse, error) {
	// 1. proto请求 → service入参（userID）转换
	userID := req.UserId

	// 2. 调用service层方法
	result, err := h.userService.GetUserInfo(ctx, userID)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("获取用户信息接口未知错误", zap.Error(err), zap.Int64("user_id", userID))
			return &userProto.GetUserInfoResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.GetUserInfoResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. service返回结果 → proto响应转换
	return &userProto.GetUserInfoResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "查询成功",
		Data: &userProto.UserInfo{
			UserId:    result.UserID,
			Username:  result.Username,
			Phone:     result.Phone,
			Avatar:    result.Avatar,
			Role:      result.Role,
			CreatedAt: result.CreatedAt,
		},
	}, nil
}

// AddAddress gRPC添加地址接口
func (h *UserHandler) AddAddress(ctx context.Context, req *userProto.AddAddressRequest) (*userProto.AddAddressResponse, error) {
	// 1. proto请求 → service入参（AddAddressParam）转换
	param := service.AddAddressParam{
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}

	// 2. 调用service层方法
	addressID, err := h.userService.AddAddress(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("添加地址接口未知错误", zap.Error(err), zap.Any("param", param))
			return &userProto.AddAddressResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.AddAddressResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. service返回结果 → proto响应转换
	return &userProto.AddAddressResponse{
		Code:      utils.ErrCodeSuccess,
		Msg:       "添加成功",
		AddressId: addressID,
	}, nil
}

// ListAddresses gRPC获取地址列表接口
func (h *UserHandler) ListAddresses(ctx context.Context, req *userProto.ListAddressesRequest) (*userProto.ListAddressesResponse, error) {
	// 1. proto请求 → service入参（userID）转换
	userID := req.UserId

	// 2. 调用service层方法
	results, err := h.userService.ListAddresses(ctx, userID)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("获取地址列表接口未知错误", zap.Error(err), zap.Int64("user_id", userID))
			return &userProto.ListAddressesResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.ListAddressesResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. service返回结果 → proto响应转换
	var protoAddrs []*userProto.Address
	for _, result := range results {
		protoAddrs = append(protoAddrs, &userProto.Address{
			AddressId: result.AddressID,
			UserId:    result.UserID,
			Receiver:  result.Receiver,
			Phone:     result.Phone,
			Province:  result.Province,
			City:      result.City,
			District:  result.District,
			Detail:    result.Detail,
			IsDefault: result.IsDefault,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		})
	}

	return &userProto.ListAddressesResponse{
		Code:      utils.ErrCodeSuccess,
		Msg:       "查询成功",
		Addresses: protoAddrs,
	}, nil
}

// UpdateUserInfo gRPC更新用户信息接口
func (h *UserHandler) UpdateUserInfo(ctx context.Context, req *userProto.UpdateUserInfoRequest) (*userProto.UpdateUserInfoResponse, error) {
	// 1. proto → service参数转换
	param := service.UpdateUserInfoParam{
		UserID:   req.UserId,
		Username: req.Username,
		Avatar:   req.Avatar,
	}

	// 2. 调用service
	err := h.userService.UpdateUserInfo(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("更新用户信息未知错误", zap.Error(err), zap.Int64("user_id", req.UserId))
			return &userProto.UpdateUserInfoResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.UpdateUserInfoResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. 返回响应
	return &userProto.UpdateUserInfoResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新成功",
	}, nil
}

// UpdateAddress gRPC更新地址接口
func (h *UserHandler) UpdateAddress(ctx context.Context, req *userProto.UpdateAddressRequest) (*userProto.UpdateAddressResponse, error) {
	// 1. proto → service参数转换
	param := service.UpdateAddressParam{
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

	// 2. 调用service
	err := h.userService.UpdateAddress(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("更新地址未知错误", zap.Error(err), zap.Any("req", req))
			return &userProto.UpdateAddressResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.UpdateAddressResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. 返回响应
	return &userProto.UpdateAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "更新成功",
	}, nil
}

// DeleteAddress gRPC删除地址接口
func (h *UserHandler) DeleteAddress(ctx context.Context, req *userProto.DeleteAddressRequest) (*userProto.DeleteAddressResponse, error) {
	// 1. proto → service参数转换
	param := service.DeleteAddressParam{
		AddressID: req.AddressId,
		UserID:    req.UserId,
	}

	// 2. 调用service
	err := h.userService.DeleteAddress(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("删除地址未知错误", zap.Error(err), zap.Any("req", req))
			return &userProto.DeleteAddressResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.DeleteAddressResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. 返回响应
	return &userProto.DeleteAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "删除成功",
	}, nil
}

// SetDefaultAddress gRPC设置默认地址接口
func (h *UserHandler) SetDefaultAddress(ctx context.Context, req *userProto.SetDefaultAddressRequest) (*userProto.SetDefaultAddressResponse, error) {
	// 1. proto → service参数转换
	param := service.SetDefaultAddressParam{
		UserID:    req.UserId,
		AddressID: req.AddressId,
	}

	// 2. 调用service
	err := h.userService.SetDefaultAddress(ctx, param)

	// 3. 错误处理
	if err != nil {
		var appErr *utils.AppError
		ok := errors.As(err, &appErr)
		if !ok {
			zap.L().Error("设置默认地址未知错误", zap.Error(err), zap.Any("req", req))
			return &userProto.SetDefaultAddressResponse{
				Code: utils.ErrCodeSystem,
				Msg:  "系统错误",
			}, nil
		}
		return &userProto.SetDefaultAddressResponse{
			Code: int32(appErr.Code),
			Msg:  appErr.Message,
		}, nil
	}

	// 4. 返回响应
	return &userProto.SetDefaultAddressResponse{
		Code: utils.ErrCodeSuccess,
		Msg:  "设置成功",
	}, nil
}
