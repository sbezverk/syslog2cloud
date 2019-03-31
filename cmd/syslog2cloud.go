package main

import (
	"context"
	"flag"
	"os"

	"github.com/knative/eventing-sources/pkg/kncloudevents"
	"github.com/sbezverk/syslog2cloud/pkg/manager"
	"github.com/sbezverk/syslog2cloud/pkg/syslog"
	"go.uber.org/zap"
)

const (
	// Number of workers waiting for a syslog message from the syslog server
	maxQueueLength = 10
	sinkURI        = "syslog2cloud.sbezverk.cisco.com"
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

	// Creating  a queue channel of maxQueueLength elements
	queue := make(chan []byte, maxQueueLength)

	ce, err := kncloudevents.NewDefaultClient(sinkURI)
	if err != nil {
		logger.Errorf("Failed to cloud events client with error: %+v", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	// Starting message manager
	if err := manager.Server(ctx, queue, logger, ce); err != nil {
		logger.Errorf("Failed to start messages manager with error: %+v", err)
		os.Exit(1)
	}
	// Starting syslog server
	if err := syslog.Server(ctx, queue, ":514", logger); err != nil {
		logger.Errorf("Failed to start UDP server with error: %+v", err)
		os.Exit(1)
	}
	defer cancel()
}
