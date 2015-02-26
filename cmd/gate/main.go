package main

import (
	"encoding/json"
	"errors"
	"flag"
	_ "net/http/pprof"
	"os"
	"strings"
	"syscall"
	"time"

	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
	"github.com/concourse/atc"
	"github.com/concourse/gate"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
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

var resourceTypes = flag.String(
	"resourceTypes",
	`[
		{"type": "archive", "image": "docker:///concourse/archive-resource" },
		{"type": "docker-image", "image": "docker:///concourse/docker-image-resource" },
		{"type": "git", "image": "docker:///concourse/git-resource" },
		{"type": "github-release", "image": "docker:///concourse/github-release-resource" },
		{"type": "s3", "image": "docker:///concourse/s3-resource" },
		{"type": "semver", "image": "docker:///concourse/semver-resource" },
		{"type": "time", "image": "docker:///concourse/time-resource" },
		{"type": "tracker", "image": "docker:///concourse/tracker-resource" }
	]`,
	"map of resource type to its rootfs",
)

var platform = flag.String(
	"platform",
	"",
	"platform that this gate is advertising",
)

var tags = flag.String(
	"tags",
	"",
	"comma separated list of tags that should be advertised",
)

func main() {
	flag.Parse()

	logger := lager.NewLogger("gate")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	if *platform == "" {
		logger.Fatal("platform-not-specified", errors.New("a -platform must be specified"))
	}

	atcEndpoint := rata.NewRequestGenerator(*atcAPIURL, atc.Routes)

	gardenClient := gclient.New(gconn.New("tcp", *gardenAddr))

	var resourceTypesNG []atc.WorkerResourceType
	err := json.Unmarshal([]byte(*resourceTypes), &resourceTypesNG)
	if err != nil {
		logger.Fatal("invalid-resource-types", err)
	}

	workerTags := []string{}
	for _, tag := range strings.Split(*tags, ",") {
		stripped := strings.TrimSpace(tag)
		if len(stripped) == 0 {
			continue
		}

		workerTags = append(workerTags, stripped)
	}

	workerToRegister := atc.Worker{
		Addr:          *gardenAddr,
		ResourceTypes: resourceTypesNG,
		Platform:      *platform,
		Tags:          workerTags,
	}

	heartbeater := gate.NewHeartbeater(
		logger,
		*heartbeatInterval,
		gardenClient,
		atcEndpoint,
		workerToRegister,
	)

	drainer := gate.NewDrainer(
		logger,
		gardenClient,
	)

	running := ifrit.Invoke(
		sigmon.New(grouper.NewParallel(nil, grouper.Members{
			{"heartbeater", heartbeater},
			{"drainer", ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
				close(ready)
				<-signals
				return drainer.Drain()
			})},
		}), syscall.SIGINT, syscall.SIGTERM),
	)

	logger.Info("started", lager.Data{
		"interval": (*heartbeatInterval).String(),
	})

	err = <-running.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
}
