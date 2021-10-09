// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/types/pipeline"
)

// PreAssembleBuild is called before setting up any containers for the build
func (c *client) PreAssembleBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("pre-assemble build %s", b.ID)

	return nil
}

// PostAssembleBuild is called after all containers have been setup
func (c *client) PostAssembleBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("post-assemble build %s", b.ID)

	return nil
}
