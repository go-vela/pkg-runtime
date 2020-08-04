// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"fmt"
	"strings"

	"github.com/go-vela/pkg-runtime/runtime/docker"
	"github.com/go-vela/pkg-runtime/runtime/kubernetes"

	"github.com/go-vela/types/constants"

	"github.com/sirupsen/logrus"
)

// Setup represents the configuration necessary for
// creating a Vela engine capable of integrating
// with a configured runtime environment.
type Setup struct {
	Driver    string
	Config    string
	Namespace string
	Volumes   []string
}

// Docker creates and returns a Vela engine capable of
// integrating with a Docker runtime environment.
func (s *Setup) Docker() (Engine, error) {
	logrus.Trace("creating docker runtime client from setup")

	// create new Docker runtime engine
	//
	// https://pkg.go.dev/github.com/go-vela/pkg-runtime/runtime/docker?tab=doc#New
	return docker.New(s.Volumes)
}

// Kubernetes creates and returns a Vela engine capable of
// integrating with a Kubernetes runtime environment.
func (s *Setup) Kubernetes() (Engine, error) {
	logrus.Trace("creating kubernetes runtime client from setup")

	// create new Kubernetes runtime engine
	//
	// https://pkg.go.dev/github.com/go-vela/pkg-runtime/runtime/kubernetes?tab=doc#New
	return kubernetes.New(s.Namespace, s.Config, s.Volumes)
}

// Validate verifies the necessary fields for the
// provided configuration are populated correctly.
func (s *Setup) Validate() error {
	logrus.Trace("validating runtime setup for client")

	// check if a runtime driver was provided
	if len(s.Driver) == 0 {
		return fmt.Errorf("no runtime driver provided in setup")
	}

	// check if the runtime driver provided is for Kubernetes
	if strings.EqualFold(s.Driver, constants.DriverKubernetes) {
		// check if a runtime namespace was provided
		if len(s.Namespace) == 0 {
			return fmt.Errorf("no runtime namespace provided in setup")
		}
	}

	// setup is valid
	return nil
}
