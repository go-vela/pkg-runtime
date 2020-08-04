// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"

	vol "github.com/go-vela/pkg-runtime/internal/volume"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// CreateVolume creates the pipeline volume.
func (c *client) CreateVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("creating volume for pipeline %s", b.ID)

	// create the volume for the pod
	//
	// This is done due to the nature of how volumes works inside
	// the pod. Each container inside the pod can access and use
	// the same volume. This allows them to share this volume
	// throughout the life of the pod. However, to keep the
	// runtime behavior consistent, Vela uses an emtpyDir volume
	// because that volume only exists for the life
	// of the pod.
	//
	// More info:
	//   * https://kubernetes.io/docs/concepts/workloads/pods/pod/
	//   * https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#Volume
	volume := v1.Volume{
		Name: b.ID,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}

	// add the volume definition to the pod spec
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#PodSpec
	c.pod.Spec.Volumes = append(c.pod.Spec.Volumes, volume)

	// check if other volumes were provided
	if len(c.volumes) > 0 {
		// iterate through all volumes provided
		for k, v := range c.volumes {
			// parse the volume provided
			_volume := vol.Parse(v)

			// add the volume to the set of pod volumes
			c.pod.Spec.Volumes = append(c.pod.Spec.Volumes, v1.Volume{
				Name: fmt.Sprintf("%s_%d", b.ID, k),
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: _volume.Source,
					},
				},
			})
		}
	}

	return nil
}

// InspectVolume inspects the pipeline volume.
func (c *client) InspectVolume(ctx context.Context, b *pipeline.Build) ([]byte, error) {
	logrus.Tracef("inspecting volume for pipeline %s", b.ID)

	// marshal the volume information from the pod
	bytes, err := json.Marshal(c.pod.Spec.Volumes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// RemoveVolume deletes the pipeline volume.
//
// Currently, this is comparable to a no-op because in Kubernetes the
// volume lives and dies with the pod it's attached to. However, Vela
// uses it to cleanup the volume definition for the pod.
func (c *client) RemoveVolume(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("removing volume for pipeline %s", b.ID)

	// remove the volume definition from the pod spec
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#PodSpec
	c.pod.Spec.Volumes = []v1.Volume{}

	return nil
}
