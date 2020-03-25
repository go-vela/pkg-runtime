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

// expected version for the Docker API
const version = "1.38"

type client struct {
	// https://godoc.org/github.com/docker/docker/client#CommonAPIClient
	Runtime docker.CommonAPIClient

	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	ctnConf *container.Config
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConf *container.HostConfig
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	netConf *network.NetworkingConfig
}

// New returns an Engine implementation that
// integrates with a Docker runtime.
func New() (*client, error) {
	// create Docker client from environment
	//
	// https://godoc.org/github.com/docker/docker/client#NewClientWithOpts
	_docker, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	// pin version to ensure we know what Docker API version we're using
	//
	// typically this would be inherited from the host environment
	// but this ensures the version of client being used
	//
	// https://godoc.org/github.com/docker/docker/client#WithVersion
	err = docker.WithVersion(version)(_docker)
	if err != nil {
		return nil, err
	}

	return &client{
		Runtime:  _docker,
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
	_docker, err := docker.NewClient("tcp://127.0.0.1:2333", version, mock, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// create the client object
	c := &client{
		Runtime: _docker,
	}

	return c, nil
}
