package conf

import (
	"X_IM/pkg/logger"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kelseyhightower/envconfig"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Redis config
const (
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

type Server struct {
}

type Config struct {
	ServiceID     string   `envconfig:"serviceId"`
	Namespace     string   `envconfig:"namespace"`
	Listen        string   `envconfig:"listen"`
	PublicAddress string   `envconfig:"publicAddress"`
	PublicPort    int      `envconfig:"publicPort"`
	Tags          []string `envconfig:"tags"`
	ConsulURL     string   `envconfig:"consulURL"`
	RedisAddrs    string   `envconfig:"redisAddrs"`
	RpcURL        string   `envconfig:"ppcURL"`
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}
	logger.Info(config)

	return &config, nil
}

func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  defaultDialTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	})

	_, err := redisDB.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return redisDB, nil
}

// InitFailoverRedis init redis with sentinels,故障转移
func InitFailoverRedis(masterName string, sentinelAddrs []string, password string, timeout time.Duration) (*redis.Client, error) {
	redisDB := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
		DialTimeout:   time.Second * 5,
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
	})

	_, err := redisDB.Ping().Result()
	if err != nil {
		logrus.Warn(err)
	}
	return redisDB, nil
}
