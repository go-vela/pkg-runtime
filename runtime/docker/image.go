// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"

	"github.com/docker/distribution/reference"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// InspectImage inspects the pipeline container image.
func (c *client) InspectImage(ctx context.Context, ctn *pipeline.Container) ([]byte, error) {
	logrus.Tracef("inspecting image for container %s", ctn.ID)

	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		return nil, err
	}

	// send API call to inspect the image
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ImageInspectWithRaw
	i, _, err := c.Runtime.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return nil, err
	}

	return []byte(i.ID + "\n"), nil
}

// parseImage is a helper function to parse
// the image for the provided container.
func parseImage(s string) (string, error) {
	logrus.Tracef("parsing image %s", s)

	// create fully qualified reference
	//
	// https://godoc.org/github.com/docker/distribution/reference#ParseNormalizedNamed
	image, err := reference.ParseNormalizedNamed(s)
	if err != nil {
		return "", err
	}

	// add latest tag to image if no tag was provided
	//
	// https://godoc.org/github.com/docker/distribution/reference#TagNameOnly
	return reference.TagNameOnly(image).String(), nil
}
