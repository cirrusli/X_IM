package services

import (
	"X_IM/logger"
	"X_IM/services/gateway"
	"X_IM/services/server"
	"context"
	"flag"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "X_IM",
		Version: version,
		Short:   "Come to see X_IM",
		Long:    "A distributed instant messaging system",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
