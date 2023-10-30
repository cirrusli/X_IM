package gateway

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/logger"
	"X_IM/naming"
	"X_IM/naming/consul"
	"X_IM/services/gateway/conf"
	"X_IM/services/gateway/serv"
	"X_IM/websocket"
	"X_IM/wire/common"
	"context"
	"github.com/spf13/cobra"
	"time"
)

type StartOptions struct {
	config   string
	protocol string
}

// NewServerStartCmd creates a new http server command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./gateway/conf.yaml", "Config file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")
	return cmd
}

// RunServerStart run http server
func RunServerStart(ctx context.Context, opts *StartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = logger.Init(logger.Settings{
		Level: "trace",
	})

	handler := &serv.Handler{
		ServiceID: config.ServiceID,
	}

	var srv x.Server
	service := &naming.DefaultService{
		ID:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
	}
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	}

	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	_ = container.Init(srv, common.SNChat, common.SNLogin)

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	// set a dialer
	container.SetDialer(serv.NewDialer(config.ServiceID))

	return container.Start()
}
