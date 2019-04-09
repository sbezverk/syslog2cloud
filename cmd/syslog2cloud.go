package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/digitalocean/captainslog"
	"github.com/sbezverk/syslog2cloud/pkg/manager"
	"github.com/sbezverk/syslog2cloud/pkg/syslog"
	"github.com/sbezverk/syslog2cloud/pkg/utils"

	corev1 "k8s.io/api/core/v1"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Number of workers waiting for a syslog message from the syslog server
	maxQueueLength = 10
	sinkURI        = "http://syslog2cloud.default.svc.k7.sbezverk.cisco.com"
)

var (
	sink      string
	namespace string
	api       string
	kind      string
	kubeCfg   string
	logger    *zap.SugaredLogger
	register  = flag.String("mode", "", "")
)

func init() {
	flag.StringVar(&sink, "sink", "", "uri to send events to.")
	flag.StringVar(&namespace, "namespace", "default", "namespace to watch events for.")
	flag.StringVar(&api, "sink-api", "v1", "API Version of sink object, defaults to v1.")
	flag.StringVar(&kind, "sink-kind", "Service", "Kind of sink object, default to Service.")
	flag.StringVar(&kubeCfg, "kubeconfig", "", "Path to kubeconfig file.")
}

func main() {

	flag.Parse()
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, err := logConfig.Build()
	if err != nil {
		log.Fatalf("unable to create logger with error: %v", err)
	}
	logger = l.Sugar()
	_, _, c, err := utils.GetClient(kubeCfg)
	if err != nil {
		log.Fatalf("unable to create client with error: %v", err)
	}
	if namespace == "" {
		logger.Fatal("no namespace provided")
	}

	if sink == "" {
		logger.Fatal("no sink provided")
	}

	// Creating URI from provided sink name and namespace
	o := &corev1.ObjectReference{
		Kind:       kind,
		APIVersion: api,
		Name:       sink,
		Namespace:  namespace,
	}

	sinkURI, err := utils.GetSinkURI(context.Background(), c, o, namespace)
	{
		logger.Fatalf("failed to get URI for sink: %s/%s with error: %+v", namespace, sink, err)
	}
	logger.Infof("URI to be used for the sink: %s", sinkURI)
	// Creating  a queue channel of maxQueueLength elements
	queue := make(chan captainslog.SyslogMsg, maxQueueLength)

	ce, err := utils.NewDefaultClient(sinkURI)
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
