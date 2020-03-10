// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"fmt"

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
}

// Docker creates and returns a Vela engine capable of
// integrating with a Docker runtime environment.
func (s *Setup) Docker() (Engine, error) {
	logrus.Trace("creating docker runtime client from setup")

	return docker.New()
}

// Kubernetes creates and returns a Vela engine capable of
// integrating with a Kubernetes runtime environment.
func (s *Setup) Kubernetes() (Engine, error) {
	logrus.Trace("creating kubernetes runtime client from setup")

	return kubernetes.New(s.Namespace, s.Config)
}

// New creates and returns a Vela engine capable of integrating
// with the configured runtime environment. Currently the
// following runtimes are supported:
//
// * docker
// * kubernetes
func New(s *Setup) (Engine, error) {
	logrus.Debug("creating runtime client from setup")

	switch s.Driver {
	case constants.DriverDocker:
		return s.Docker()
	case constants.DriverKubernetes:
		return s.Kubernetes()
	default:
		return nil, fmt.Errorf("invalid runtime driver: %s", s.Driver)
	}
}
