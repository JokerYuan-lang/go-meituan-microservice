package utils

import "fmt"

const (
	ErrCodeSuccess = 0     // 成功
	ErrCodeParam   = 10001 // 参数错误
	ErrCodeAuth    = 10002 // 鉴权错误
	ErrCodeDB      = 10003 // 数据库错误
	ErrCodeBiz     = 10004 // 业务错误
	ErrCodeSystem  = 99999 // 系统错误
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *AppError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", err.Code, err.Message)
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewParamError(message string) *AppError {
	return NewAppError(ErrCodeParam, message)
}

func NewAuthError(message string) *AppError {
	return NewAppError(ErrCodeAuth, message)
}

func NewDBError(message string) *AppError {
	return NewAppError(ErrCodeDB, message)
}

func NewBizError(message string) *AppError {
	return NewAppError(ErrCodeBiz, message)
}

func NewSystemError(message string) *AppError {
	return NewAppError(ErrCodeSystem, message)
}
