package main

import (
	"X_IM/examples/mock"
	"X_IM/logger"
	"context"
	"flag"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "test",
		Version: version,
		Short:   "server",
	}
	ctx := context.Background()

	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatalln("Could not run command")
	}
}
