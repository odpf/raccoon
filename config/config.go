package config

import (
	"github.com/spf13/viper"
)

var loaded bool

func Load() {
	if loaded {
		return
	}
	loaded = true
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("kafka_queue_size", "100000")
	viper.SetDefault("kafka_flush_interval", "1000")
	viper.SetConfigName("application")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.SetConfigType("yaml")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	AppPort()
	LogLevel()
	NewKafkaConfig()
	BufferConfigLoader()
}
