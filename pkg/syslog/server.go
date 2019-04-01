package syslog

import (
	"context"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

// maxBufferSize specifies the maximum size of the syslog
// message received over UDP
const maxBufferSize = 1024

var (
	validPriority = regexp.MustCompile(`^<[0-9]+>`)
)

// Server receives and  logs syslog messages
func Server(ctx context.Context, queue chan []byte, address string, logger *zap.SugaredLogger) (err error) {
	logger.Infof("Starting listening for UPD packets...")
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		logger.Errorf("ListenPacket failed with error: %+v", err)
		return
	}
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}
			// get the timestamp of recevied message
			ts := time.Now().Format("Jan 02 15:04:05") + " "
			logger.Infof("Syslog Message from: %s content: %s", addr.String(), string(buffer[:n]))
			go transferSyslogMsg(buffer[:n], ts, addr.String(), queue, logger)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Server's context has been cancelled, exiting...")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return
}

func transferSyslogMsg(b []byte, timeStamp, h string, queue chan []byte, logger *zap.SugaredLogger) {
	// Sending received message for further processing
	pri := validPriority.FindString(string(b))
	content := strings.TrimPrefix(string(b), pri)
	host := h[:strings.LastIndex(h, ":")] + " "
	tag := uuid.New().String()[:8]

	var msg []byte
	msg = append(msg, pri...)
	msg = append(msg, timeStamp...)
	msg = append(msg, host...)
	msg = append(msg, tag...)
	msg = append(msg, content...)
	queue <- msg
}
