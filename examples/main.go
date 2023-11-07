package main

import (
	"X_IM/examples/mock"
	"X_IM/pkg/logger"
	"context"
	"flag"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "",
		Version: version,
		Short:   "there is mock test",
	}
	ctx := context.Background()

	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatalln("Could not run command")
	}
}
