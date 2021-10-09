// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/types/pipeline"
)

// SetupBuild prepares the pipeline build.
// This is a no-op for docker.
func (c *client) SetupBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("setting up for build %s", b.ID)

	return nil
}

// AssembleBuild finalizes pipeline build setup.
// This is a no-op for docker.
func (c *client) AssembleBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("assembling build %s", b.ID)

	return nil
}

// RemoveBuild deletes (kill, remove) the pipeline build metadata.
// This is a no-op for docker.
func (c *client) RemoveBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("removing build %s", b.ID)

	return nil
}
