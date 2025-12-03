package main

import (
	"flag"
	"fmt"
	"net"

	userProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/user/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/config"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/db"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/kafka"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/middleware"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/redis"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var configPath = flag.String("config", "config.yaml", "配置文件路径")

func main() {
	//初始化配置
	config.InitConfig(*configPath)
	defer zap.L().Sync()

	//初始化服务
	db.InitMysql()
	redis.InitRedis()
	kafka.InitKafkaProducer()

	//启动grpc服务

	grpcPort := config.Cfg.GRPC.UserPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zap.L().Fatal("gRPC监听失败", zap.Error(err), zap.Int("port", grpcPort))
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCJwtMiddleware()),
	)

	userProto.RegisterUserServiceServer(grpcServer)

}
