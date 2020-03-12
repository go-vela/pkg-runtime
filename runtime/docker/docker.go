// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/go-vela/pkg-runtime/runtime/docker/testdata/mock"
	"github.com/sirupsen/logrus"
)

const dockerVersion = "1.38"

type client struct {
	Runtime *docker.Client

	ctnConf  *container.Config
	hostConf *container.HostConfig
	netConf  *network.NetworkingConfig
}

// New returns an Engine implementation that
// integrates with a Docker runtime.
func New() (*client, error) {
	// create Docker client from environment
	r, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	// pin version to prevent "client version <version> is too new." errors
	// typically this would be inherited from the host env but this will ensure
	// we know what version of the Docker API we're using
	err = docker.WithVersion(dockerVersion)(r)
	if err != nil {
		return nil, err
	}

	return &client{
		Runtime:  r,
		ctnConf:  new(container.Config),
		hostConf: new(container.HostConfig),
		netConf:  new(network.NetworkingConfig),
	}, nil
}

// NewMock returns an Engine implementation that
// integrates with a mock Docker runtime.
//
// This function is intended for running tests only.
func NewMock() (*client, error) {
	// create mock client
	mock := mock.Client(mock.Router)

	// create Docker client from the mock client
	r, err := docker.NewClient("tcp://127.0.0.1:2333", dockerVersion, mock, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// create the client object
	c := &client{
		Runtime: r,
	}

	return c, nil
}
