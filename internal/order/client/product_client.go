package client

import (
	productProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/product/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ProductClient productProto.ProductServiceClient // 全局商品服务客户端

// InitProductClient 初始化商品服务gRPC客户端
func InitProductClient() {
	// 商品服务地址
	addr := "localhost:50052"

	// 连接商品服务
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.L().Fatal("连接商品服务失败", zap.String("addr", addr), zap.Error(err))
	}

	// 创建客户端
	ProductClient = productProto.NewProductServiceClient(conn)
	zap.L().Info("商品服务客户端初始化成功", zap.String("addr", addr))
}
