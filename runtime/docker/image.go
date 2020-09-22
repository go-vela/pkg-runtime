// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/go-vela/pkg-runtime/internal/image"
	"github.com/go-vela/types/constants"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// CreateImage creates the pipeline container image.
func (c *client) CreateImage(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("creating image for container %s", ctn.ID)

	// parse image from container
	//
	// https://pkg.go.dev/github.com/go-vela/pkg-runtime/internal/image#ParseWithError
	_image, err := image.ParseWithError(ctn.Image)
	if err != nil {
		return err
	}

	// create options for pulling image
	//
	// https://godoc.org/github.com/docker/docker/api/types#ImagePullOptions
	opts := types.ImagePullOptions{}

	// send API call to pull the image for the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ImagePull
	reader, err := c.docker.ImagePull(ctx, _image, opts)
	if err != nil {
		return err
	}

	defer reader.Close()

	// copy output from image pull to standard output
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}

	return nil
}

// InspectImage inspects the pipeline container image.
func (c *client) InspectImage(ctx context.Context, ctn *pipeline.Container) ([]byte, error) {
	logrus.Tracef("inspecting image for container %s", ctn.ID)

	// parse image from container
	//
	// https://pkg.go.dev/github.com/go-vela/pkg-runtime/internal/image#ParseWithError
	_image, err := image.ParseWithError(ctn.Image)
	if err != nil {
		return nil, err
	}

	// check if the container pull policy is on start
	if strings.EqualFold(ctn.Pull, constants.PullOnStart) {
		return []byte(fmt.Sprintf("skipped for container %s due to pull policy %s\n", ctn.ID, ctn.Pull)), nil
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
