package utils

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func BcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("bcrypt加密密码失败", zap.String("password", password), zap.Error(err))
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		zap.L().Warn("密码验证失败", zap.Error(err))
		return false
	}
	return true
}
