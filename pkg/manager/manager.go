package manager

import (
	"context"

	"go.uber.org/zap"
)

// Server receives and  logs syslog messages
func Server(ctx context.Context, queue chan []byte, logger *zap.SugaredLogger) (err error) {
	logger.Infof("Starting Messages Manager ..")
Exit:
	for {
		select {
		case <-ctx.Done():
			logger.Warn("Messages manager's context has been cancelled, exiting...")
			err = ctx.Err()
			break Exit
		case msg := <-queue:
			logger.Infof("Syslog Message:%s", string(msg))
		}
	}
	return
}
