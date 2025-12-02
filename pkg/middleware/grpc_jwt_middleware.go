package middleware

import (
	"context"
	"strings"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GRPCJwtMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		noAuthMethods := map[string]bool{
			"/user.UserService/Register": true,
			"/user.UserService/Login":    true,
		}
		if noAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Warn("gRPC请求未携带Metadata", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "未携带鉴权信息")
		}
		authHeaders := md.Get("Authorization")
		if len(authHeaders) == 0 {
			zap.L().Warn("gRPC请求未携带Authorization头", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "未携带Token")
		}
		authHeader := authHeaders[0]
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			zap.L().Warn("Authorization头格式错误", zap.String("header", authHeader), zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "Token格式错误（应为Bearer <token>）")
		}
		tokenStr := tokenParts[1]
		//解析tokenStr
		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			zap.L().Warn("JWT Token解析失败", zap.String("token", tokenStr), zap.Error(err), zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "Token无效："+err.Error())
		}
		ctx = context.WithValue(ctx, "token", claims)
		return handler(ctx, req)
	}
}
