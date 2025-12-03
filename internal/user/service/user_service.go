package service

import (
	"context"
	"strconv"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 全局参数校验器（校验领域模型的入参）
var validate = validator.New()

// 定义service层的入参结构体（不依赖proto，纯领域层定义）
type RegisterParam struct {
	Username string `validate:"required,min=2,max=32"`          // 用户名2-32位
	Password string `validate:"required,min=6,max=20"`          // 密码6-20位
	Phone    string `validate:"required,regexp=^1[3-9]\\d{9}$"` // 手机号正则
}

type LoginParam struct {
	Phone    string `validate:"required,regexp=^1[3-9]\\d{9}$"` // 手机号正则
	Password string `validate:"required,min=6,max=20"`          // 密码6-20位
}

type UpdateUserInfoParam struct {
	UserID   int64  `validate:"required,gt=0"`
	Username string `validate:"omitempty,min=2,max=32"` // 可选，更新时传
	Avatar   string `validate:"omitempty,url"`          // 头像URL格式
}

type AddAddressParam struct {
	UserID    int64  `validate:"required,gt=0"`
	Receiver  string `validate:"required,min=2,max=32"`
	Phone     string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Province  string `validate:"required,min=2"`
	City      string `validate:"required,min=2"`
	District  string `validate:"required,min=2"`
	Detail    string `validate:"required,min=5"`
	IsDefault bool   `validate:"required"`
}

type UpdateAddressParam struct {
	AddressID int64  `validate:"required,gt=0"`
	UserID    int64  `validate:"required,gt=0"`
	Receiver  string `validate:"required,min=2,max=32"`
	Phone     string `validate:"required,regexp=^1[3-9]\\d{9}$"`
	Province  string `validate:"required,min=2"`
	City      string `validate:"required,min=2"`
	District  string `validate:"required,min=2"`
	Detail    string `validate:"required,min=5"`
	IsDefault bool   `validate:"required"`
}

type DeleteAddressParam struct {
	AddressID int64 `validate:"required,gt=0"`
	UserID    int64 `validate:"required,gt=0"`
}

type SetDefaultAddressParam struct {
	UserID    int64 `validate:"required,gt=0"`
	AddressID int64 `validate:"required,gt=0"`
}

// 定义service层的返回结构体（纯领域层，不依赖proto）
type LoginResult struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type AddressResult struct {
	AddressID int64  `json:"address_id"`
	UserID    int64  `json:"user_id"`
	Receiver  string `json:"receiver"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserInfoResult struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// UserService 业务逻辑层接口（入参/返回值均为领域层类型）
type UserService interface {
	Register(ctx context.Context, param RegisterParam) (int64, string, error) // 返回userID、token、错误
	Login(ctx context.Context, param LoginParam) (LoginResult, error)
	GetUserInfo(ctx context.Context, userID int64) (UserInfoResult, error)
	UpdateUserInfo(ctx context.Context, param UpdateUserInfoParam) error
	AddAddress(ctx context.Context, param AddAddressParam) (int64, error) // 返回addressID、错误
	ListAddresses(ctx context.Context, userID int64) ([]AddressResult, error)
	UpdateAddress(ctx context.Context, param UpdateAddressParam) error
	DeleteAddress(ctx context.Context, param DeleteAddressParam) error
	SetDefaultAddress(ctx context.Context, param SetDefaultAddressParam) error
}

// userService 接口实现
type userService struct {
	userRepo    repo.UserRepo // 依赖repo接口，不依赖具体实现
	addressRepo repo.AddressRepo
}

// NewUserService 创建UserService实例（依赖注入repo）
func NewUserService(userRepo repo.UserRepo, addressRepo repo.AddressRepo) UserService {
	return &userService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
	}
}

// Register 用户注册（业务逻辑：参数校验→手机号去重→创建用户→生成Token）
func (s *userService) Register(ctx context.Context, param RegisterParam) (int64, string, error) {
	// 1. 参数校验（领域层自己校验，不依赖proto的validate）
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("注册参数校验失败", zap.Any("param", param), zap.Error(err))
		return 0, "", utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 校验手机号是否已注册
	existUser, err := s.userRepo.GetUserByPhone(ctx, param.Phone)
	if err != nil {
		return 0, "", err // repo返回的是AppError，直接向上抛
	}
	if existUser != nil {
		return 0, "", utils.NewBizError("手机号已注册")
	}

	// 3. 转换为领域模型（model）
	user := &model.User{
		Username: param.Username,
		Password: param.Password, // 原始密码，GORM钩子会自动bcrypt加密
		Phone:    param.Phone,
		Role:     "user",
	}

	// 4. 调用repo创建用户
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return 0, "", err
	}

	// 5. 生成JWT Token
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(user.UserID, 10),
		Username: user.Username,
		Phone:    user.Phone,
		Role:     user.Role,
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成注册Token失败", zap.Int64("user_id", user.UserID), zap.Error(err))
		return user.UserID, "", utils.NewSystemError("注册成功，但生成Token失败")
	}

	zap.L().Info("用户注册成功", zap.Int64("user_id", user.UserID), zap.String("phone", param.Phone))
	return user.UserID, token, nil
}

// Login 用户登录（业务逻辑：参数校验→查询用户→密码验证→生成Token）
func (s *userService) Login(ctx context.Context, param LoginParam) (LoginResult, error) {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("登录参数校验失败", zap.Any("param", param), zap.Error(err))
		return LoginResult{}, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用repo查询用户
	user, err := s.userRepo.GetUserByPhone(ctx, param.Phone)
	if err != nil {
		return LoginResult{}, err
	}
	if user == nil {
		return LoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 3. bcrypt密码验证
	if !utils.CheckPasswordHash(param.Password, user.Password) {
		return LoginResult{}, utils.NewBizError("手机号或密码错误")
	}

	// 4. 生成JWT Token
	jwtClaims := &utils.UserClaims{
		UserID:   strconv.FormatInt(user.UserID, 10),
		Username: user.Username,
		Phone:    user.Phone,
		Role:     user.Role,
	}
	token, err := utils.GenerateToken(jwtClaims)
	if err != nil {
		zap.L().Error("生成登录Token失败", zap.Int64("user_id", user.UserID), zap.Error(err))
		return LoginResult{}, utils.NewSystemError("登录失败，生成Token失败")
	}

	// 5. 转换为领域层返回结果
	result := LoginResult{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
	}

	zap.L().Info("用户登录成功", zap.Int64("user_id", user.UserID), zap.String("phone", param.Phone))
	return result, nil
}

// GetUserInfo 获取用户信息（业务逻辑：校验用户ID→查询用户→转换结果）
func (s *userService) GetUserInfo(ctx context.Context, userID int64) (UserInfoResult, error) {
	// 1. 参数校验
	if userID <= 0 {
		return UserInfoResult{}, utils.NewParamError("用户ID不能为空且必须大于0")
	}

	// 2. 调用repo查询用户
	user, err := s.userRepo.GetUserByUserID(ctx, userID)
	if err != nil {
		return UserInfoResult{}, err
	}
	if user == nil {
		return UserInfoResult{}, utils.NewBizError("用户不存在")
	}

	// 3. 转换为领域层返回结果
	result := UserInfoResult{
		UserID:    user.UserID,
		Username:  user.Username,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// AddAddress 添加收货地址（业务逻辑：参数校验→创建地址→设置默认地址）
func (s *userService) AddAddress(ctx context.Context, param AddAddressParam) (int64, error) {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("添加地址参数校验失败", zap.Any("param", param), zap.Error(err))
		return 0, utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 转换为领域模型
	addr := &model.Address{
		UserID:    param.UserID,
		Receiver:  param.Receiver,
		Phone:     param.Phone,
		Province:  param.Province,
		City:      param.City,
		District:  param.District,
		Detail:    param.Detail,
		IsDefault: param.IsDefault,
	}

	// 3. 调用repo创建地址
	if err := s.addressRepo.CreateAddress(ctx, addr); err != nil {
		return 0, err
	}

	// 4. 如果是默认地址，更新其他地址为非默认
	if param.IsDefault {
		if err := s.addressRepo.UpdateDefaultAddress(ctx, param.UserID, addr.AddressID); err != nil {
			zap.L().Warn("设置默认地址失败", zap.Int64("address_id", addr.AddressID), zap.Error(err))
			// 不影响地址创建，仅日志警告
		}
	}

	zap.L().Info("添加收货地址成功", zap.Int64("address_id", addr.AddressID), zap.Int64("user_id", param.UserID))
	return addr.AddressID, nil
}

// ListAddresses 获取地址列表（业务逻辑：校验用户ID→查询地址→转换结果）
func (s *userService) ListAddresses(ctx context.Context, userID int64) ([]AddressResult, error) {
	// 1. 参数校验
	if userID <= 0 {
		return nil, utils.NewParamError("用户ID不能为空且必须大于0")
	}

	// 2. 调用repo查询地址列表
	addrs, err := s.addressRepo.ListAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. 转换为领域层返回结果
	var results []AddressResult
	for _, addr := range addrs {
		results = append(results, AddressResult{
			AddressID: addr.AddressID,
			UserID:    addr.UserID,
			Receiver:  addr.Receiver,
			Phone:     addr.Phone,
			Province:  addr.Province,
			City:      addr.City,
			District:  addr.District,
			Detail:    addr.Detail,
			IsDefault: addr.IsDefault,
			CreatedAt: addr.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: addr.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return results, nil
}

// UpdateUserInfo 更新用户信息
func (s *userService) UpdateUserInfo(ctx context.Context, param UpdateUserInfoParam) error {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("更新用户信息参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 查询用户是否存在
	user, err := s.userRepo.GetUserByUserID(ctx, param.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBizError("用户不存在")
	}

	// 3. 只更新传入的非空字段
	if param.Username != "" {
		user.Username = param.Username
	}
	if param.Avatar != "" {
		user.Avatar = param.Avatar
	}

	// 4. 调用repo更新
	return s.userRepo.UpdateUser(ctx, user)
}

// UpdateAddress 更新收货地址
func (s *userService) UpdateAddress(ctx context.Context, param UpdateAddressParam) error {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("更新地址参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 查询地址是否存在且属于该用户
	addr, err := s.addressRepo.GetAddressByID(ctx, param.AddressID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != param.UserID {
		return utils.NewBizError("地址不存在或不属于该用户")
	}

	// 3. 更新字段
	addr.Receiver = param.Receiver
	addr.Phone = param.Phone
	addr.Province = param.Province
	addr.City = param.City
	addr.District = param.District
	addr.Detail = param.Detail
	addr.IsDefault = param.IsDefault

	// 4. 调用repo更新
	if err := s.addressRepo.UpdateAddress(ctx, addr); err != nil {
		return err
	}

	// 5. 如果设置为默认地址，同步更新其他地址
	if param.IsDefault {
		if err := s.addressRepo.UpdateDefaultAddress(ctx, param.UserID, param.AddressID); err != nil {
			zap.L().Warn("更新默认地址失败", zap.Int64("address_id", param.AddressID), zap.Error(err))
		}
	}

	return nil
}

// DeleteAddress 删除收货地址
func (s *userService) DeleteAddress(ctx context.Context, param DeleteAddressParam) error {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("删除地址参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用repo删除（软删除）
	return s.addressRepo.DeleteAddress(ctx, param.AddressID, param.UserID)
}

// SetDefaultAddress 设置默认地址
func (s *userService) SetDefaultAddress(ctx context.Context, param SetDefaultAddressParam) error {
	// 1. 参数校验
	if err := validate.Struct(param); err != nil {
		zap.L().Warn("设置默认地址参数校验失败", zap.Any("param", param), zap.Error(err))
		return utils.NewParamError("参数错误：" + err.Error())
	}

	// 2. 调用repo设置默认地址（事务保证）
	return s.addressRepo.UpdateDefaultAddress(ctx, param.UserID, param.AddressID)
}
