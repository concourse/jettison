package main

import (
	"flag"
	_ "net/http/pprof"
	"os"

	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
	"github.com/concourse/jettison"
	"github.com/pivotal-golang/lager"
	"github.com/xoebus/zest"
)

var gardenAddr = flag.String(
	"gardenAddr",
	"127.0.0.1:7777",
	"garden API host:port",
)

var yellerAPIKey = flag.String(
	"yellerAPIKey",
	"",
	"API token to output error logs to Yeller",
)
var yellerEnvironment = flag.String(
	"yellerEnvironment",
	"development",
	"environment label for Yeller",
)

func main() {
	flag.Parse()

	logger := lager.NewLogger("jettison")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	if *yellerAPIKey != "" {
		yellerSink := zest.NewYellerSink(*yellerAPIKey, *yellerEnvironment)
		logger.RegisterSink(yellerSink)
	}

	gardenClient := gclient.New(gconn.New("tcp", *gardenAddr))

	drainer := jettison.NewDrainer(
		logger,
		gardenClient,
	)

	err := drainer.Drain()
	if err != nil {
		logger.Fatal("draining-failed", err)
	}

	logger.Info("drained")
}
