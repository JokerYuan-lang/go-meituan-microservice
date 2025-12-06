package client

import (
	orderProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/order/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var OrderClient orderProto.OrderServiceClient

func InitOrderClient() {
	addr := "localhost:50054"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.L().Fatal("连接订单服务", zap.String("addr", addr), zap.Error(err))
	}

	//创建客户端
	OrderClient = orderProto.NewOrderServiceClient(conn)
	zap.L().Info("订单服务客户端初始化成功", zap.String("addr", addr))
}
