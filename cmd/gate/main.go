package main

import (
	"flag"
	_ "net/http/pprof"
	"os"
	"time"

	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
	"github.com/concourse/atc"
	"github.com/concourse/gate"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/rata"
)

var heartbeatInterval = flag.Duration(
	"heartbeatInterval",
	30*time.Second,
	"interval on which to register with the ATC.",
)

var gardenAddr = flag.String(
	"gardenAddr",
	"127.0.0.1:7777",
	"garden API host:port",
)

var atcAPIURL = flag.String(
	"atcAPIURL",
	"http://127.0.0.1:8080",
	"ATC API URL to register with",
)

func main() {
	flag.Parse()

	logger := lager.NewLogger("gate")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	atcEndpoint := rata.NewRequestGenerator(*atcAPIURL, atc.Routes)

	gardenClient := gclient.New(gconn.New("tcp", *gardenAddr))

	running := ifrit.Invoke(gate.NewHeartbeater(logger, *gardenAddr, *heartbeatInterval, gardenClient, atcEndpoint))

	logger.Info("started", lager.Data{
		"interval": (*heartbeatInterval).String(),
	})

	err := <-running.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
}
