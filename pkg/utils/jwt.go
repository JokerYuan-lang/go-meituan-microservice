package utils

import (
	"errors"
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/config"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(claims *UserClaims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(config.Cfg.Jwt.Expire)))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())  // 签发时间
	claims.NotBefore = jwt.NewNumericDate(time.Now()) // 生效时间（立即生效）
	//创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//生成Token字符串
	tokenStr, err := token.SignedString([]byte(config.Cfg.Jwt.Secret))
	if err != nil {
		zap.L().Error("生成JWT Token失败", zap.Any("claims", claims), zap.Error(err))
		return "", err
	}
	return tokenStr, nil
}

func ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			zap.L().Error("JWT签名方法不匹配", zap.String("method", token.Header["alg"].(string)))
			return nil, errors.New("签名方法不合法")
		}
		return []byte(config.Cfg.Jwt.Secret), nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				zap.L().Warn("JWT Token已过期", zap.String("token", tokenString))
				return nil, errors.New("token已过期")
			}
		}
		zap.L().Error("解析JWT Token失败", zap.String("token", tokenString), zap.Error(err))
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	zap.L().Error("JWT Token无效", zap.String("token", tokenString))
	return nil, errors.New("token无效")
}
