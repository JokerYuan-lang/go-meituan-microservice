package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/handler"
	merchantProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/repo/model"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/merchant/service"
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
	// 初始化配置和依赖
	config.InitConfig(*configPath)
	defer zap.L().Sync()
	db.InitMysql()
	if err := db.Mysql.AutoMigrate(&model.Merchant{}); err != nil {
		zap.L().Fatal("商家表迁移失败", zap.Error(err))
	}
	redis.InitRedis()
	kafka.InitKafkaProducer()
	defer func() {
		if kafka.Producer != nil {
			_ = kafka.Producer.Close()
		}
	}()

	// 依赖注入
	merchantRepo := repo.NewMerchantRepo()
	merchantService := service.NewMerchantService(merchantRepo)
	merchantHandler := handler.NewMerchantHandler(merchantService)

	// 启动gRPC服务
	grpcPort := config.Cfg.GRPC.MerchantPort // 配置文件中添加商家服务端口（如50053）
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zap.L().Fatal("商家服务gRPC监听失败", zap.Error(err), zap.Int("port", grpcPort))
	}
	defer func() {
		_ = listen.Close()
	}()

	// 创建gRPC服务器（添加JWT鉴权）
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCJwtMiddleware()),
	)
	merchantProto.RegisterMerchantServiceServer(grpcServer, merchantHandler)

	zap.L().Info("商家服务启动成功", zap.String("addr", fmt.Sprintf("localhost:%d", grpcPort)))

	// 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		zap.L().Info("商家服务开始关闭...")
		grpcServer.GracefulStop()
		zap.L().Info("商家服务已关闭")
	}()

	// 启动服务
	if err = grpcServer.Serve(listen); err != nil {
		zap.L().Fatal("商家服务启动失败", zap.Error(err))
	}
}
