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

type config struct {
	// specifies a list of privileged images to use for the Docker client
	Images []string
	// specifies a list of host volumes to use for the Docker client
	Volumes []string
}

type client struct {
	config *config
	// https://godoc.org/github.com/docker/docker/client#CommonAPIClient
	Docker docker.CommonAPIClient
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	ContainerConfig *container.Config
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	HostConfig *container.HostConfig
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	NetworkConfig *network.NetworkingConfig
}

// New returns an Engine implementation that
// integrates with a Docker runtime.
//
// nolint: golint // ignore returning unexported client
func New(opts ...ClientOpt) (*client, error) {
	// create new Docker client
	c := new(client)

	// create new fields
	c.config = new(config)
	c.ContainerConfig = new(container.Config)
	c.HostConfig = new(container.HostConfig)
	c.NetworkConfig = new(network.NetworkingConfig)

	// apply all provided configuration options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// create new Docker client from environment
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

	// set the Docker client in the runtime client
	c.Docker = _docker

	return c, nil
}

// NewMock returns an Engine implementation that
// integrates with a mock Docker runtime.
//
// This function is intended for running tests only.
//
// nolint: golint // ignore returning unexported client
func NewMock(opts ...ClientOpt) (*client, error) {
	// create new Docker runtime client
	c, err := New(opts...)
	if err != nil {
		return nil, err
	}

	// create Docker client from the mock client
	//
	// https://pkg.go.dev/github.com/go-vela/mock/docker#New
	_docker, err := mock.New()
	if err != nil {
		return nil, err
	}

	// set the Docker client in the runtime client
	c.Docker = _docker

	return c, nil
}
