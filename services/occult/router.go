package occult

import (
	"X_IM/naming"
	"X_IM/naming/consul"
	log "X_IM/pkg/logger"
	"X_IM/pkg/middleware"
	"X_IM/services/occult/conf"
	"X_IM/services/occult/database"
	"X_IM/services/occult/handler"
	"X_IM/wire/common"
	"context"
	"fmt"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"hash/crc32"
	"net/http"

	"github.com/kataras/iris/v12"
)

type ServerStartOptions struct {
	config string
}

func newApp(serviceHandler *handler.ServiceHandler) *iris.Application {
	app := iris.Default()
	limiter := middleware.NewRateLimiter(middleware.SlidingWindow, 5, 5)
	app.Use(func(ctx iris.Context) {
		h := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Next()
		}))
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
	})

	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})
	messageAPI := app.Party("/api/:app/message")
	{
		messageAPI.Post("/user", serviceHandler.InsertUserMessage)
		messageAPI.Post("/group", serviceHandler.InsertGroupMessage)
		messageAPI.Post("/ack", serviceHandler.MessageACK)
	}

	groupAPI := app.Party("/api/:app/group")
	{
		groupAPI.Get("/:id", serviceHandler.GroupGet)
		groupAPI.Post("", serviceHandler.GroupCreate)
		groupAPI.Post("/member", serviceHandler.GroupJoin)
		groupAPI.Delete("/member", serviceHandler.GroupQuit)
		groupAPI.Get("/members/:id", serviceHandler.GroupMembers)
	}

	offlineAPI := app.Party("/api/:app/offline")
	{
		offlineAPI.Use(iris.Compression)
		offlineAPI.Post("/index", serviceHandler.GetOfflineMessageIndex)
		offlineAPI.Post("/content", serviceHandler.GetOfflineMessageContent)
	}
	return app
}

// NewServerStartCmd creates a new http logic command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "occult",
		Short: "Start a RPC service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./occult/conf.yaml", "Config file")
	return cmd
}

// RunServerStart run http server
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = log.Init(log.Settings{
		Level:    config.LogLevel,
		Filename: "./data/occult.log",
	})

	// database.Init
	var (
		baseDB    *gorm.DB
		messageDB *gorm.DB
	)
	baseDB, err = database.InitDB(config.Driver, config.BaseDB)
	if err != nil {
		return err
	}
	messageDB, err = database.InitDB(config.Driver, config.MessageDB)
	if err != nil {
		return err
	}

	_ = baseDB.AutoMigrate(&database.Group{}, &database.GroupMember{})
	_ = messageDB.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})

	if config.NodeID == 0 {
		config.NodeID = int64(HashCode(config.ServiceID))
	}
	idgen, err := database.NewIDGenerator(config.NodeID)
	if err != nil {
		return err
	}

	rdb, err := conf.InitRedis(config.RedisAddrs, config.RedisPass)
	if err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	_ = ns.Register(&naming.DefaultService{
		ID:       config.ServiceID,
		Name:     common.SNService, // service name
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: "http",
		Tags:     config.Tags,
		Meta: map[string]string{
			consul.KeyHealthURL: fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.PublicPort),
		},
	})
	log.Infoln("consul health URL is: ",
		fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.PublicPort))
	defer func() {
		_ = ns.Deregister(config.ServiceID)
	}()
	serviceHandler := handler.ServiceHandler{
		BaseDB:    baseDB,
		MessageDB: messageDB,
		IDGen:     idgen,
		Cache:     rdb,
	}

	ac := conf.MakeAccessLog()
	defer func(ac *accesslog.AccessLog) {
		_ = ac.Close()
	}(ac)

	app := newApp(&serviceHandler)
	app.UseRouter(ac.Handler)
	app.UseRouter(setAllowedResponses)

	// Start HTTP server
	return app.Listen(config.Listen, iris.WithOptimizations)
}

func setAllowedResponses(ctx iris.Context) {
	// Indicate that the Server can send JSON, XML, YAML and MessagePack for this request.
	ctx.Negotiation().JSON().Protobuf().MsgPack()
	// Add more, allowed by the server format of responses, mime types here...

	// If client is missing an "Accept: " header then default it to JSON.
	ctx.Negotiation().Accept.JSON()

	ctx.Next()
}

func HashCode(key string) uint32 {
	hash32 := crc32.NewIEEE()
	_, _ = hash32.Write([]byte(key))
	return hash32.Sum32() % 1000
}
