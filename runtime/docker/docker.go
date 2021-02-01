// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"

	mock "github.com/go-vela/mock/docker"
)

// expected version for the Docker API.
const version = "1.40"

type client struct {
	// https://godoc.org/github.com/docker/docker/client#CommonAPIClient
	docker docker.CommonAPIClient

	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	ctnConf *container.Config
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConf *container.HostConfig
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	netConf *network.NetworkingConfig

	// set of host volumes to mount into every container
	volumes []string
	// set of images that are allowed to run in privileged mode
	privilegedImages []string
}

// New returns an Engine implementation that
// integrates with a Docker runtime.
//
// nolint: golint // ignore returning unexported client
func New(_volumes, _privilegedImages []string) (*client, error) {
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
	_ = docker.WithVersion(version)(_docker)

	return &client{
		docker:           _docker,
		ctnConf:          new(container.Config),
		hostConf:         new(container.HostConfig),
		netConf:          new(network.NetworkingConfig),
		volumes:          _volumes,
		privilegedImages: _privilegedImages,
	}, nil
}

// NewMock returns an Engine implementation that
// integrates with a mock Docker runtime.
//
// This function is intended for running tests only.
//
// nolint: golint // ignore returning unexported client
func NewMock() (*client, error) {
	// create Docker client from the mock client
	_docker, _ := mock.New()

	// create the client object
	c := &client{
		docker:           _docker,
		privilegedImages: []string{"target/vela-git"},
	}

	return c, nil
}
