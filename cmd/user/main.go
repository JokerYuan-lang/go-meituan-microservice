package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/handler"
	userProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/user/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/user/service"
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
	_ = config.InitConfig(*configPath)
	defer zap.L().Sync()

	//初始化服务
	db.InitMysql()
	//if err := db.Mysql.AutoMigrate(&model.User{}, &model.Address{}); err != nil {
	//	zap.L().Fatal("数据库表迁移失败", zap.Error(err))
	//}
	redis.InitRedis()
	kafka.InitKafkaProducer()
	defer func() {
		if kafka.Producer != nil {
			_ = kafka.Producer.Close()
		}
	}()
	userRepo := repo.NewUserRepo()
	addressRepo := repo.NewAddressRepo()
	userService := service.NewUserService(userRepo, addressRepo)
	userHandler := handler.NewUserHandler(userService)
	//启动grpc服务
	grpcPort := config.Cfg.GRPC.UserPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zap.L().Fatal("gRPC监听失败", zap.Error(err), zap.Int("port", grpcPort))
	}
	defer func() {
		_ = lis.Close()
	}()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCJwtMiddleware()),
	)

	userProto.RegisterUserServiceServer(grpcServer, userHandler)
	zap.L().Info("用户服务启动成功", zap.String("addr", fmt.Sprintf("localhost:%d", grpcPort)))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		zap.L().Info("用户服务开始关闭...")
		grpcServer.GracefulStop()
		zap.L().Info("用户服务已关闭")
	}()

	// 6. 启动gRPC服务
	if err = grpcServer.Serve(lis); err != nil {
		zap.L().Fatal("gRPC服务启动失败", zap.Error(err))
	}
}
