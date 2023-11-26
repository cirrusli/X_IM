package router

import (
	"X_IM/internal/router/api"
	"X_IM/internal/router/conf"
	"X_IM/internal/router/ip"
	"X_IM/pkg/logger"
	"X_IM/pkg/middleware"
	"X_IM/pkg/naming"
	"X_IM/pkg/naming/consul"
	"X_IM/pkg/wire/common"
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"net/http"
	"path"

	"github.com/kataras/iris/v12"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	confPath = "../internal/router/conf.yaml"
	logPath  = "./data/router.log"
	dataPath = "../internal/router/data"
)

type ServerStartOptions struct {
	config string
	data   string
}

// NewServerStartCmd creates a new http server command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "router",
		Short: "Start a router server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", confPath, "Config file")
	cmd.PersistentFlags().StringVarP(&opts.data, "data", "d", dataPath, "data path")
	return cmd
}

// RunServerStart run http server
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = logger.Init(logger.Settings{
		Level:    "info",
		Filename: logPath,
	})

	mappings, err := conf.LoadMapping(path.Join(opts.data, "mapping.json"))
	if err != nil {
		return err
	}
	logrus.Infof("load mappings - %v", mappings)
	regions, err := conf.LoadRegions(path.Join(opts.data, "regions.json"))
	if err != nil {
		return err
	}
	logrus.Infof("load regions - %v", regions)

	region, err := ip.NewSearcher(path.Join(opts.data, "ip2region.db"))
	if err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	_ = ns.Register(&naming.DefaultService{
		ID:       config.ServiceID,
		Name:     common.SNRouter,
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
	router := api.RouterApi{
		Naming:   ns,
		IpRegion: region,
		Config: conf.Router{
			Mapping: mappings,
			Regions: regions,
		},
	}

	app := iris.Default()
	// 限流
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

	routerAPI := app.Party("/api/lookup")
	{
		routerAPI.Get("/:token", router.Lookup)
	}

	// Start HTTP server
	return app.Listen(config.Listen, iris.WithOptimizations)
}
