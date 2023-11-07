package conf

import (
	x "X_IM/pkg"
	"X_IM/pkg/logger"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	ServiceID       string
	Listen          string `default:":8005"`
	MonitorPort     int    `default:"8006"`
	PublicAddress   string
	PublicPort      int `default:"8005"`
	Tags            []string
	Zone            string `default:"zone_ali_03"`
	ConsulURL       string
	RedisAddrs      string
	OccultURL       string
	LogLevel        string `default:"DEBUG"`
	MessageGPool    int    `default:"5000"`
	ConnectionGPool int    `default:"500"`
}

func (c Config) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config
	err := envconfig.Process("x_im", &config)
	if err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(err)
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	if config.ServiceID == "" {
		localIP := x.GetLocalIP()
		config.ServiceID = fmt.Sprintf("server_%s", strings.ReplaceAll(localIP, ".", ""))
	}
	if config.PublicAddress == "" {
		config.PublicAddress = x.GetLocalIP()
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
