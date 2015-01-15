package gate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/concourse/atc"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/rata"
)

type heartbeater struct {
	logger lager.Logger

	addrToRegister string
	interval       time.Duration

	gardenClient garden.Client
	atcEndpoint  *rata.RequestGenerator
}

func NewHeartbeater(
	logger lager.Logger,
	addrToRegister string,
	interval time.Duration,
	gardenClient garden.Client,
	atcEndpoint *rata.RequestGenerator,
) ifrit.Runner {
	return &heartbeater{
		logger: logger,

		addrToRegister: addrToRegister,
		interval:       interval,

		gardenClient: gardenClient,
		atcEndpoint:  atcEndpoint,
	}
}

func (heartbeater *heartbeater) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	for !heartbeater.register(heartbeater.logger.Session("register")) {
		select {
		case <-time.After(time.Second):
		case <-signals:
			return nil
		}
	}

	close(ready)

	for {
		select {
		case <-signals:
			return nil

		case <-time.After(heartbeater.interval):
			heartbeater.register(heartbeater.logger.Session("heartbeat"))
		}
	}
}

func (heartbeater *heartbeater) register(logger lager.Logger) bool {
	logger.Info("start")
	defer logger.Info("done")

	containers, err := heartbeater.gardenClient.Containers(nil)
	if err != nil {
		logger.Error("failed-to-fetch-containers", err)
		return false
	}

	registration := atc.Worker{
		Addr:             heartbeater.addrToRegister,
		ActiveContainers: len(containers),
	}

	payload, err := json.Marshal(registration)
	if err != nil {
		logger.Error("failed-to-marshal-registration", err)
		return false
	}

	request, err := heartbeater.atcEndpoint.CreateRequest(atc.RegisterWorker, nil, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("failed-to-construct-request", err)
		return false
	}

	request.URL.RawQuery = url.Values{
		"ttl": []string{heartbeater.ttl().String()},
	}.Encode()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Error("failed-to-register", err)
		return false
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Error("bad-response", nil, lager.Data{
			"status-code": response.StatusCode,
		})

		return false
	}

	return true
}

func (heartbeater *heartbeater) ttl() time.Duration {
	return heartbeater.interval * 2
}
