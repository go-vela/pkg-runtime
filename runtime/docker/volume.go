// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// CreateVolume creates the pipeline volume.
func (c *client) CreateVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("creating volume for pipeline %s", b.ID)

	// create host configuration
	c.hostConf = hostConfig(b.ID)

	// create options for creating volume
	//
	// https://godoc.org/github.com/docker/docker/api/types/volume#VolumeCreateBody
	opts := volume.VolumeCreateBody{
		Name:   b.ID,
		Driver: "local",
	}

	// send API call to create the volume
	//
	// https://godoc.org/github.com/docker/docker/client#Client.VolumeCreate
	_, err := c.Runtime.VolumeCreate(ctx, opts)
	if err != nil {
		return err
	}

	return nil
}

// InspectVolume inspects the pipeline volume.
func (c *client) InspectVolume(ctx context.Context, b *pipeline.Build) ([]byte, error) {
	logrus.Tracef("inspecting volume for pipeline %s", b.ID)

	// send API call to inspect the volume
	//
	// https://godoc.org/github.com/docker/docker/client#Client.VolumeInspect
	v, err := c.Runtime.VolumeInspect(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	return []byte(v.Name + "\n"), nil
}

// RemoveVolume deletes the pipeline volume.
func (c *client) RemoveVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("removing volume for pipeline %s", b.ID)

	// send API call to remove the volume
	//
	// https://godoc.org/github.com/docker/docker/client#Client.VolumeRemove
	err := c.Runtime.VolumeRemove(ctx, b.ID, true)
	if err != nil {
		return err
	}

	return nil
}

// hostConfig is a helper function to generate
// the host config with volume specification for a container.
func hostConfig(id string) *container.HostConfig {
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	return &container.HostConfig{
		// https://godoc.org/github.com/docker/docker/api/types/container#LogConfig
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: false,
		// https://godoc.org/github.com/docker/docker/api/types/mount#Mount
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: id,
				Target: "/home",
			},
		},
	}
}
