package service

import (
	"context"

	userProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/user/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo"
	"github.com/go-playground/validator/v10"
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

}

func (u *userService) Login(ctx context.Context, req *userProto.LoginRequest) (*userProto.LoginResponse, error) {

}

func (u *userService) GetUserInfo(ctx context.Context, req *userProto.GetUserInfoRequest) (*userProto.GetUserInfoResponse, error) {

}

func (u *userService) UpdateUserInfo(ctx context.Context, req *userProto.UpdateUserInfoRequest) (*userProto.UpdateUserInfoResponse, error) {

}

func (u *userService) AddAddress(ctx context.Context, req *userProto.AddAddressRequest) (*userProto.AddAddressResponse, error) {

}

func (u *userService) ListAddresses(ctx context.Context, req *userProto.ListAddressesRequest) (*userProto.ListAddressesResponse, error) {

}
