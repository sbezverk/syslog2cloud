package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/digitalocean/captainslog"

	"go.uber.org/zap"
)

const (
	eventType = "dev.knative.syslog.message"
)

// Server receives and  logs syslog messages
func Server(ctx context.Context, queue chan captainslog.SyslogMsg, logger *zap.SugaredLogger, ce client.Client) error {
	var err error
	logger.Infof("Starting Messages Manager ..")
	go func() {
		for {
			select {
			case msg := <-queue:
				logger.Infof("Syslog Message from:%s", msg.Host)
				go transmitSyslogMsg(ce, msg, logger)
			case <-ctx.Done():
				logger.Warn("Messages manager's context has been cancelled, exiting...")
				err = ctx.Err()
				return
			}
		}
	}()

	return err
}

// Retry logic?
func transmitSyslogMsg(ce client.Client, msg captainslog.SyslogMsg, logger *zap.SugaredLogger) {

	ce2Send, err := cloudEventFrom(msg)
	if err != nil {
		logger.Errorf("Failed to parse syslog message into a cloud event with error: %+v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	cloudEvent, err := ce.Send(ctx, *ce2Send)
	if err != nil {
		logger.Errorf("Failed to send cloud event with error: %+v", err)
		return
	}
	logger.Infof("Succeeded to send cloud event: %+v", cloudEvent)
}

func cloudEventFrom(syslogMsg captainslog.SyslogMsg) (*cloudevents.Event, error) {
	url := types.ParseURLRef("/Syslog/" + syslogMsg.Host)
	if url == nil {
		return nil, fmt.Errorf("ParseURLRef returned nil for: %s", "/syslog/"+syslogMsg.Host)
	}
	return &cloudevents.Event{
		Context: cloudevents.EventContextV02{
			Type:   eventType,
			Source: *url,
			Time:   &types.Timestamp{Time: syslogMsg.Time},
		}.AsV02(),
		Data: syslogMsg.Content,
	}, nil
}
