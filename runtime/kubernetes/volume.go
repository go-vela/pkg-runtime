// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/json"

	"k8s.io/api/core/v1"

	"github.com/go-vela/types/pipeline"
)

// CreateVolume creates the pipeline volume.
func (c *client) CreateVolume(ctx context.Context, b *pipeline.Build) error {
	volume := v1.Volume{
		Name: b.ID,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}

	c.Pod.Spec.Volumes = append(c.Pod.Spec.Volumes, volume)

	return nil
}

// InspectVolume inspects the pipeline volume.
func (c *client) InspectVolume(ctx context.Context, b *pipeline.Build) ([]byte, error) {
	bytes, err := json.Marshal(c.Pod.Spec.Volumes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// RemoveVolume deletes the pipeline volume.
//
// TODO: research this
//
// currently this is a no-op because in Kubernetes the
// volume lives and dies with the pod it's attached to
func (c *client) RemoveVolume(ctx context.Context, b *pipeline.Build) error {
	return nil
}
