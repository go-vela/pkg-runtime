// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"fmt"

	"github.com/go-vela/types/constants"

	"github.com/sirupsen/logrus"
)

// New creates and returns a Vela engine capable of integrating
// with the configured runtime environment. Currently the
// following runtimes are supported:
//
// * docker
// * kubernetes
func New(s *Setup) (Engine, error) {
	// validate the setup being provided
	err := s.Validate()
	if err != nil {
		return nil, err
	}

	logrus.Debug("creating runtime client from setup")
	// process the runtime driver being provided
	switch s.Driver {
	case constants.DriverDocker:
		// handle the Docker runtime driver being provided
		return s.Docker()
	case constants.DriverKubernetes:
		// handle the Kubernetes runtime driver being provided
		return s.Kubernetes()
	default:
		// handle an invalid runtime driver being provided
		return nil, fmt.Errorf("invalid runtime driver provided in setup: %s", s.Driver)
	}
}
