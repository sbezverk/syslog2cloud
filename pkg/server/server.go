package server

import (
	"context"
	"net"

	"go.uber.org/zap"
)

// maxBufferSize specifies the size of the buffers that
// are used to temporarily hold data from the UDP packets
// that we receive.
const maxBufferSize = 1024

// Server receives and  logs syslog messages
func Server(ctx context.Context, address string, logger *zap.SugaredLogger) (err error) {
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

			logger.Infof("packet-received: bytes=%d from=%s\n",
				n, addr.String())
			logger.Infof("Syslog Message: %s", string(buffer[:n]))
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
