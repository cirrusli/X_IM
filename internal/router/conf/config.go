package conf

import (
	"X_IM/pkg/logger"
	"github.com/bytedance/sonic"

	"github.com/kelseyhightower/envconfig"

	"github.com/spf13/viper"
)

// Config Config
type Config struct {
	Listen        string `default:":8100"`
	ConsulURL     string `default:"localhost:8500"`
	LogLevel      string `default:"INFO"`
	ServiceID     string
	PublicAddress string
	PublicPort    int
	Tags          []string `default:"router"`
}

func (c Config) String() string {
	bts, _ := sonic.Marshal(c)
	return string(bts)
}

// Init InitConfig
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

	return &config, nil
}
