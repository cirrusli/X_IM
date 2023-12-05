package conf

import (
	"X_IM/pkg"
	"X_IM/pkg/logger"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/kelseyhightower/envconfig"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/spf13/viper"
)

const logPath = "./data/access.log"

type Config struct {
	ServiceID     string
	NodeID        int64
	Listen        string `default:":8080"`
	PublicAddress string
	PublicPort    int `default:"8080"`
	Tags          []string
	ConsulURL     string
	RedisAddrs    string
	RedisPass     string
	Driver        string `default:"mysql"`
	BaseDB        string
	MessageDB     string
	LogLevel      string `default:"INFO"`
}

func (c Config) String() string {
	bts, _ := sonic.Marshal(c)
	return string(bts)
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config
	// 之前envconfig放在viper解析后如果字段(如LogLevel)从配置文件中读取,
	// 由于有default标签,会覆盖配置文件中的值
	// 需要envconfig是因为配置文件中没有的字段(如Driver),可以直接用结构体字段的default
	err := envconfig.Process("x_im", &config)
	if err != nil {
		return nil, err
	}
	logger.Infoln("envconfig: ", config)
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(err)
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}
	logger.Infoln("viper: ", config)

	if config.ServiceID == "" {
		localIP := pkg.GetLocalIP()
		config.ServiceID = fmt.Sprintf("occult_%s", strings.ReplaceAll(localIP, ".", ""))
		arr := strings.Split(localIP, ".")
		if len(arr) == 4 {
			suffix, _ := strconv.Atoi(arr[3])
			config.NodeID = int64(suffix)
		}
	}
	if config.PublicAddress == "" {
		config.PublicAddress = pkg.GetLocalIP()
	}
	logger.Info("last: ", config)
	return &config, nil
}

func MakeAccessLog() *accesslog.AccessLog {
	// Initialize a new access log middleware.
	ac := accesslog.File(logPath)
	// Remove this line to disable logging to console:
	ac.AddOutput(os.Stdout)

	// The default configuration:
	ac.Delim = '|'
	ac.TimeFormat = time.DateTime
	ac.Async = false
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = false
	ac.BytesSent = false
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = false
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler

	// Default line format if formatter is missing:
	// Time|Latency|Code|Method|Path|IP|Path Params Query Fields|Bytes Received|Bytes Sent|Request|Response|
	//
	// Set Custom Formatter:
	// ac.SetFormatter(&accesslog.JSON{
	// 	Indent:    "  ",
	// 	HumanTime: true,
	// })
	// ac.SetFormatter(&accesslog.CSV{})
	// ac.SetFormatter(&accesslog.Template{Text: "{{.Code}}"})

	return ac
}
