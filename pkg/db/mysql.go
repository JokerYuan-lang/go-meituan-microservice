package db

import (
	"fmt"

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
}
