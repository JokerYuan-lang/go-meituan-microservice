package kafka

import (
	"github.com/IBM/sarama"
	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/config"
	"go.uber.org/zap"
)

var Producer sarama.SyncProducer

func InitKafkaProducer() {
	cfg := config.Cfg.Kafka
	//生产者配置
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll //等待所有分区确认
	producerConfig.Producer.Retry.Max = 3                    //重试次数
	producerConfig.Producer.Return.Successes = true          //成功交付的消息会返回给调用者
	producerConfig.Version = sarama.V2_0_0_0

	//创建生产者
	producer, err := sarama.NewSyncProducer(cfg.Brokers, producerConfig)
	if err != nil {
		zap.L().Fatal("kafka生产者初始化失败", zap.Error(err))
	}
	Producer = producer
	zap.L().Info("kafka生产者初始化成功")
}

func SendMessage(topic string, key string, value string) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}
	partition, offset, err := Producer.SendMessage(msg)
	return partition, offset, err
}
