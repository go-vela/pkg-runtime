// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"

	"github.com/docker/distribution/reference"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

const imagePatch = `
{
  "spec": {
    "containers": [
      {
        "name": "%s",
        "image": "%s"
      }
    ]
  }
}
`

// InspectImage inspects the pipeline container image.
func (c *client) InspectImage(ctx context.Context, ctn *pipeline.Container) ([]byte, error) {
	logrus.Tracef("inspecting image for container %s", ctn.ID)

	return nil, nil
}

// parseImage is a helper function to parse
// the image for the provided container.
func parseImage(s string) (string, error) {
	logrus.Tracef("parsing image %s", s)

	// create fully qualified reference
	image, err := reference.ParseNormalizedNamed(s)
	if err != nil {
		return "", err
	}

	// add latest tag to image if no tag was provided
	return reference.TagNameOnly(image).String(), nil
}
