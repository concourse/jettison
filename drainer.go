package jettison

import (
	"github.com/cloudfoundry-incubator/garden"
	"github.com/hashicorp/go-multierror"
	"github.com/pivotal-golang/lager"
)

type Drainer struct {
	logger lager.Logger

	client garden.Client
}

func NewDrainer(
	logger lager.Logger,
	client garden.Client,
) *Drainer {
	return &Drainer{
		logger: logger,
		client: client,
	}
}

func (d *Drainer) Drain() error {
	containers, err := d.client.Containers(garden.Properties{
		"concourse:ephemeral": "true",
	})
	if err != nil {
		return err
	}

	var result error

	for _, container := range containers {
		handle := container.Handle()
		err := d.client.Destroy(handle)
		if err != nil {
			result = multierror.Append(result, err)

			d.logger.Error("destroying-container", err, lager.Data{
				"handle": handle,
			})
		}
	}

	return result
}
