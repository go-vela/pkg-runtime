// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"fmt"

	"github.com/go-vela/pkg-runtime/runtime"
	"github.com/go-vela/pkg-runtime/runtime/docker"
	"github.com/go-vela/pkg-runtime/runtime/kubernetes"

	"github.com/go-vela/types/constants"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func setupRuntime(c *cli.Context) (runtime.Engine, error) {
	logrus.Debug("Creating runtime client from CLI configuration")

	switch c.String("runtime.driver") {
	case constants.DriverDocker:
		return setupDocker(c)
	case constants.DriverKubernetes:
		return setupKubernetes(c)
	default:
		return nil, fmt.Errorf("invalid runtime driver: %s", c.String("runtime.driver"))
	}
}

// helper function to setup the Docker runtime from the CLI arguments.
func setupDocker(c *cli.Context) (runtime.Engine, error) {
	logrus.Tracef("Creating %s runtime client from CLI configuration", constants.DriverDocker)
	return docker.New()
}

// helper function to setup the Docker runtime from the CLI arguments.
func setupKubernetes(c *cli.Context) (runtime.Engine, error) {
	logrus.Tracef("Creating %s runtime client from CLI configuration", constants.DriverKubernetes)
	return kubernetes.New(c.String("runtime.path"))
}
