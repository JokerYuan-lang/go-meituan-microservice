package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/handler"
	productProto "github.com/JokerYuan-lang/go-meituan-microservice/internal/product/proto"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/repo"
	"github.com/JokerYuan-lang/go-meituan-microservice/internal/product/service"
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
	_ = config.InitConfig(*configPath)
	defer zap.L().Sync()
	db.InitMysql()
	// 迁移创建商品表
	//if err := db.Mysql.AutoMigrate(&model.Product{}); err != nil {
	//	zap.L().Fatal("商品表迁移失败", zap.Error(err))
	//}

	redis.InitRedis()
	kafka.InitKafkaProducer()
	defer func() {
		if kafka.Producer != nil {
			_ = kafka.Producer.Close()
		}
	}()

	// 依赖注入
	productRepo := repo.NewProductRepo()
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// 启动gRPC服务
	grpcPort := config.Cfg.GRPC.ProductPort
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zap.L().Fatal("商品服务gRPC监听失败", zap.Error(err), zap.Int("port", grpcPort))
	}
	defer func() {
		_ = listen.Close()
	}()

	// 创建gRPC服务器（添加JWT鉴权）
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCJwtMiddleware()),
	)
	productProto.RegisterProductServiceServer(grpcServer, productHandler)

	zap.L().Info("商品服务启动成功", zap.String("addr", fmt.Sprintf("localhost:%d", grpcPort)))

	// 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		zap.L().Info("商品服务开始关闭...")
		grpcServer.GracefulStop()
		zap.L().Info("商品服务已关闭")
	}()

	// 启动服务
	if err = grpcServer.Serve(listen); err != nil {
		zap.L().Fatal("商品服务启动失败", zap.Error(err))
	}
}
