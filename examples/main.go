package main

import (
	"X_IM/examples/benchmark"
	"X_IM/examples/mock"
	"X_IM/examples/ut"
	"X_IM/pkg/logger"
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
		Short:   "test",
	}
	ctx := context.Background()

	root.AddCommand(ut.NewEchoCmd(ctx))
	root.AddCommand(benchmark.NewBenchmarkCmd(ctx))
	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatalln("Could not run command")
	}
}
