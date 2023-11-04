package logic

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/logger"
	"X_IM/naming"
	"X_IM/naming/consul"
	"X_IM/services/logic/conf"
	"X_IM/services/logic/handler"
	"X_IM/services/logic/serv"
	"X_IM/storage"
	"X_IM/tcp"
	"X_IM/wire/common"
	"context"
	"github.com/spf13/cobra"
)

type StartOptions struct {
	config      string
	serviceName string
}

// NewServerStartCmd creates a new http logic command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "logic",
		Short: "Start a logic server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config,
		"config", "c", "./logic/conf.yaml", "Config file")

	cmd.PersistentFlags().StringVarP(&opts.serviceName,
		"serviceName", "s", "chat", "defined a service name,option is login or chat")

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

	r := x.NewRouter()
	// login
	loginHandler := handler.NewLoginHandler()
	r.Handle(common.CommandLoginSignIn, loginHandler.DoLogin)
	r.Handle(common.CommandLoginSignOut, loginHandler.DoLogout)

	rdb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}
	cache := storage.NewRedisStorage(rdb)
	servHandler := serv.NewServHandler(r, cache)

	service := &naming.DefaultService{
		ID:       config.ServiceID,
		Name:     opts.serviceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(common.ProtocolTCP),
		Tags:     config.Tags,
	}
	srv := tcp.NewServer(config.Listen, service)

	srv.SetReadWait(x.DefaultReadWait)
	srv.SetAcceptor(servHandler)
	srv.SetMessageListener(servHandler)
	srv.SetStateListener(servHandler)

	if err := container.Init(srv); err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	return container.Start()
}
