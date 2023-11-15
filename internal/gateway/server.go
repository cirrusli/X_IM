package gateway

import (
	x "X_IM"
	"X_IM/internal/gateway/conf"
	"X_IM/internal/gateway/serv"
	"X_IM/pkg/container"
	"X_IM/pkg/logger"
	"X_IM/pkg/naming"
	"X_IM/pkg/naming/consul"
	"X_IM/pkg/tcp"
	"X_IM/pkg/websocket"
	"X_IM/pkg/wire/common"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	_ "net/http/pprof"
	"time"
)

// StartOptions is the options for start command
type StartOptions struct {
	config   string
	protocol string
	route    string
}

const (
	confWS    = "./gateway/ws.yaml"
	confTCP   = "./gateway/tcp.yaml"
	routePath = "./gateway/route.json"
	protocol  = "ws" //如果没有在命令行指定的话，就用这个默认值
	logPath   = "./data/gateway.log"

	ReadWait = 2 * time.Minute
)

// NewServerStartCmd creates a new http logic server command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", confWS, "Config file")
	cmd.PersistentFlags().StringVarP(&opts.route, "route", "r", routePath, "route file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", protocol, "protocol of ws or tcp")
	return cmd
}

// RunServerStart run http logic server
func RunServerStart(ctx context.Context, opts *StartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = logger.Init(logger.Settings{
		Level:    "trace",
		Filename: logPath,
	})

	handler := &serv.Handler{
		ServiceID: config.ServiceID,
		AppSecret: config.AppSecret,
	}
	meta := make(map[string]string)
	meta[consul.KeyHealthURL] = fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort)
	meta["domain"] = config.Domain

	var srv x.Server
	service := &naming.DefaultService{
		ID:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
		Meta:     meta,
	}
	srvOpts := []x.ServerOption{
		x.WithConnectionGPool(config.ConnectionGPool), x.WithMessageGPool(config.MessageGPool),
	}
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service, srvOpts...)
	} else if opts.protocol == "tcp" {
		srv = tcp.NewServer(config.Listen, service, srvOpts...)
	}

	srv.SetReadWait(ReadWait)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	_ = container.Init(srv, common.SNChat, common.SNLogin)
	container.EnableMonitor(fmt.Sprintf(":%d", config.MonitorPort))

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	// set a dialer
	container.SetDialer(serv.NewDialer(config.ServiceID))
	// use routeSelector
	selector, err := serv.NewRouteSelector(opts.route)
	if err != nil {
		return err
	}
	container.SetSelector(selector)
	return container.Start()
}
