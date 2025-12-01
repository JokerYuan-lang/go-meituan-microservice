package db

import (
	"fmt"
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/config"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var Mysql *gorm.DB

func InitMysql() {
	cfg := config.Cfg.MySQL
	mysqlDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	db, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		zap.L().Fatal("MySQL连接失败", zap.Error(err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("获取SQL DB失败", zap.Error(err))
	}

	sqlDB.SetMaxIdleConns(10)               //最大空闲连接数
	sqlDB.SetMaxOpenConns(100)              //最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour * 2) //连接最大生命周期

	Mysql = db
	zap.L().Info("Mysql初始化成功")
}
