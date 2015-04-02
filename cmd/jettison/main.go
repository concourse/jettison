package main

import (
	"flag"
	_ "net/http/pprof"
	"os"

	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
	"github.com/concourse/jettison"
	"github.com/pivotal-golang/lager"
)

var gardenAddr = flag.String(
	"gardenAddr",
	"127.0.0.1:7777",
	"garden API host:port",
)

func main() {
	flag.Parse()

	logger := lager.NewLogger("jettison")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

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
