// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"

	"github.com/go-vela/pkg-runtime/internal/image"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// InspectImage inspects the pipeline container image.
func (c *client) InspectImage(ctx context.Context, ctn *pipeline.Container) ([]byte, error) {
	logrus.Tracef("inspecting image for container %s", ctn.ID)

	// parse image from container
	_image, err := image.ParseWithError(ctn.Image)
	if err != nil {
		return nil, err
	}

	// send API call to inspect the image
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ImageInspectWithRaw
	i, _, err := c.docker.ImageInspectWithRaw(ctx, _image)
	if err != nil {
		return nil, err
	}

	return []byte(i.ID + "\n"), nil
}
