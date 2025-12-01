package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// 全局配置变量

var Cfg *Config

// 配置结构体（对应配置文件）

type Config struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
	Kafka KafkaConfig `mapstructure:"kafka"`
	ES    ESConfig    `mapstructure:"es"`
	GRPC  GRPCConfig  `mapstructure:"grpc"`
	Log   LogConfig   `mapstructure:"log"`
	Jwt   JwtConfig   `mapstructure:"jwt"`
}

// MySQL配置

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// Redis配置

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Kafka配置

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

// ES配置

type ESConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// GRPC配置

type GRPCConfig struct {
	UserPort     int `mapstructure:"user_port"`
	ProductPort  int `mapstructure:"product_port"`
	OrderPort    int `mapstructure:"order_port"`
	MerchantPort int `mapstructure:"merchant_port"`
	RiderPort    int `mapstructure:"rider_port"`
}

// 日志配置

type LogConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

func InitConfig(configPath string) error {
	viper.SetConfigFile(filepath.Clean(configPath))
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("viper read config failed, err:%v", err)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("viper unmarshal config failed, err:%v", err)
	}
	zap.L().Info("配置初始化成功")
	return nil
}
