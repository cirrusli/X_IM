package logic

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/naming"
	"X_IM/naming/consul"
	"X_IM/pkg/logger"
	"X_IM/pkg/middleware"
	"X_IM/services/logic/conf"
	"X_IM/services/logic/handler"
	"X_IM/services/logic/restful"
	"X_IM/services/logic/serv"
	"X_IM/storage"
	"X_IM/tcp"
	"X_IM/wire/common"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"strings"
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
		Level:    config.LogLevel,
		Filename: "./data/server.log",
	})

	var groupService restful.Group
	var messageService restful.Message
	if strings.TrimSpace(config.OccultURL) != "" {
		groupService = restful.NewGroupService(config.OccultURL)
		messageService = restful.NewMessageService(config.OccultURL)
	} else {
		srvRecord := &resty.SRVRecord{
			Domain:  "consul",
			Service: common.SNService,
		}
		groupService = restful.NewGroupServiceWithSRV("http", srvRecord)
		messageService = restful.NewMessageServiceWithSRV("http", srvRecord)
	}

	r := x.NewRouter()
	r.Use(middleware.Recover())

	// login
	loginHandler := handler.NewLoginHandler()
	r.Handle(common.CommandLoginSignIn, loginHandler.DoLogin)
	r.Handle(common.CommandLoginSignOut, loginHandler.DoLogout)
	// talk
	chatHandler := handler.NewChatHandler(messageService, groupService)
	r.Handle(common.CommandChatUserTalk, chatHandler.DoSingleTalk)
	r.Handle(common.CommandChatGroupTalk, chatHandler.DoGroupTalk)
	r.Handle(common.CommandChatTalkAck, chatHandler.DoTalkAck)
	// group
	groupHandler := handler.NewGroupHandler(groupService)
	r.Handle(common.CommandGroupCreate, groupHandler.DoCreate)
	r.Handle(common.CommandGroupJoin, groupHandler.DoJoin)
	r.Handle(common.CommandGroupQuit, groupHandler.DoQuit)
	r.Handle(common.CommandGroupDetail, groupHandler.DoDetail)

	// offline
	offlineHandler := handler.NewOfflineHandler(messageService)
	r.Handle(common.CommandOfflineIndex, offlineHandler.DoSyncIndex)
	r.Handle(common.CommandOfflineContent, offlineHandler.DoSyncContent)

	rdb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}
	cache := storage.NewRedisStorage(rdb)
	servHandler := serv.NewServHandler(r, cache)

	meta := make(map[string]string)
	meta[consul.KeyHealthURL] = fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort)
	meta["zone"] = config.Zone

	service := &naming.DefaultService{
		ID:       config.ServiceID,
		Name:     opts.serviceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(common.ProtocolTCP),
		Tags:     config.Tags,
		Meta:     meta,
	}
	srvOpts := []x.ServerOption{
		x.WithConnectionGPool(config.ConnectionGPool), x.WithMessageGPool(config.MessageGPool),
	}
	srv := tcp.NewServer(config.Listen, service, srvOpts...)

	srv.SetReadWait(x.DefaultReadWait)
	srv.SetAcceptor(servHandler)
	srv.SetMessageListener(servHandler)
	srv.SetStateListener(servHandler)

	if err := container.Init(srv); err != nil {
		return err
	}
	container.EnableMonitor(fmt.Sprintf(":%d", config.MonitorPort))

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	return container.Start()
}
