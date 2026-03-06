package server

import (
	"errors"

	controller "fiber-boilerplate/pkg/controllers"
)

type Dependencies struct {
	HealthController *controller.HealthController
}

func (d Dependencies) Validate() error {
	if d.HealthController == nil {
		return errors.New("server dependency HealthController is required")
	}

	return nil
}
