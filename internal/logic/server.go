package logic

import (
	"X_IM/internal/logic/client"
	"X_IM/internal/logic/conf"
	"X_IM/internal/logic/handler"
	"X_IM/internal/logic/server"
	"X_IM/pkg/container"
	"X_IM/pkg/logger"
	"X_IM/pkg/middleware"
	"X_IM/pkg/naming"
	"X_IM/pkg/naming/consul"
	"X_IM/pkg/storage"
	"X_IM/pkg/tcp"
	"X_IM/pkg/wire/common"
	"X_IM/pkg/x"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"strings"
)

const (
	//启动不同server时,注意service name也需要一起更改
	confChat  = "../internal/logic/chat.yaml"
	confLogin = "../internal/logic/login.yaml"
	logPath   = "./data/server.log"
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
		"config", "c", confChat, "Config file")

	cmd.PersistentFlags().StringVarP(&opts.serviceName,
		"serviceName", "s", common.SNChat, "defined a service name,option is login or chat")

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
		Filename: logPath,
	})

	var groupService client.Group
	var messageService client.Message
	if strings.TrimSpace(config.OccultURL) != "" {
		groupService = client.NewGroupService(config.OccultURL)
		messageService = client.NewMessageService(config.OccultURL)
	} else {
		srvRecord := &resty.SRVRecord{
			Domain:  "consul",
			Service: common.SNService,
		}
		groupService = client.NewGroupServiceWithSRV("http", srvRecord)
		messageService = client.NewMessageServiceWithSRV("http", srvRecord)
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

	rdb, err := conf.InitRedis(config.RedisAddrs, config.RedisPass)
	if err != nil {
		return err
	}
	cache := storage.NewRedisStorage(rdb)
	//cache:=storage.NewRedisClusterStorage(rdb)

	servHandler := server.NewServHandler(r, cache)

	meta := make(map[string]string)
	meta[consul.KeyHealthURL] = fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort)
	logger.Infoln("consul health URL is: ",
		fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort))
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
		x.WithConnectionGPool(config.ConnectionGPool),
		x.WithMessageGPool(config.MessageGPool),
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
