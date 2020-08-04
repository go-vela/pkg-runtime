// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"

	vol "github.com/go-vela/pkg-runtime/internal/volume"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// CreateVolume creates the pipeline volume.
func (c *client) CreateVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("creating volume for pipeline %s", b.ID)

	// create host configuration
	c.hostConf = hostConfig(b.ID, c.volumes)

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
	_, err := c.docker.VolumeCreate(ctx, opts)
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
	v, err := c.docker.VolumeInspect(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	// convert volume type Volume to bytes with pretty print
	//
	// https://godoc.org/github.com/docker/docker/api/types#Volume
	volume, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return nil, err
	}

	// add new line to end of bytes
	return append(volume, "\n"...), nil
}

// RemoveVolume deletes the pipeline volume.
func (c *client) RemoveVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("removing volume for pipeline %s", b.ID)

	// send API call to remove the volume
	//
	// https://godoc.org/github.com/docker/docker/client#Client.VolumeRemove
	err := c.docker.VolumeRemove(ctx, b.ID, true)
	if err != nil {
		return err
	}

	return nil
}

// hostConfig is a helper function to generate
// the host config with volume specification for a container.
func hostConfig(id string, volumes []string) *container.HostConfig {
	logrus.Tracef("creating mount for default volume %s", id)

	// create default mount for pipeline volume
	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: id,
			Target: "/vela",
		},
	}

	// check if other volumes were provided
	if len(volumes) > 0 {
		// iterate through all volumes provided
		for _, v := range volumes {
			logrus.Tracef("creating mount for volume %s", v)

			// parse the volume provided
			_volume, err := vol.ParseWithError(v)
			if err != nil {
				logrus.Error(err)
			}

			// add the volume to the set of mounts
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   _volume.Source,
				Target:   _volume.Destination,
				ReadOnly: _volume.AccessMode == "ro",
			})
		}
	}

	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	return &container.HostConfig{
		// https://godoc.org/github.com/docker/docker/api/types/container#LogConfig
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: false,
		// https://godoc.org/github.com/docker/docker/api/types/mount#Mount
		Mounts: mounts,
	}
}
