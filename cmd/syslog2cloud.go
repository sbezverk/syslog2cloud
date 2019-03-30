package main

import (
	"context"
	"flag"
	"os"

	"github.com/sbezverk/syslog2cloud/pkg/server"
	"go.uber.org/zap"
)

var (
	logger   *zap.SugaredLogger
	register = flag.String("mode", "", "")
)

func init() {
	// Setting up logger
	l, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	logger = l.Sugar()
}

func main() {

	server.Server(context.Background(), ":514", logger)

}
