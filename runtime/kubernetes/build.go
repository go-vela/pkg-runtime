// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/types/pipeline"
)

// SetupBuild prepares the pod metadata for the pipeline build.
func (c *client) SetupBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("setting up for build %s", b.ID)

	// create the object metadata for the pod
	//
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1?tab=doc#ObjectMeta
	c.Pod.ObjectMeta = metav1.ObjectMeta{
		Name:   b.ID,
		Labels: map[string]string{"pipeline": b.ID},
	}

	// create the restart policy for the pod
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#RestartPolicy
	c.Pod.Spec.RestartPolicy = v1.RestartPolicyNever

	return nil
}

// AssembleBuild finalizes the pipeline build setup.
// This creates the pod in kubernetes for the pipeline build.
// After creation, image is the only container field we can edit in kubernetes.
// So, all environment, volume, and other container metadata must be setup
// before running AssembleBuild.
func (c *client) AssembleBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("assembling build %s", b.ID)

	logrus.Infof("creating pod %s", c.Pod.ObjectMeta.Name)
	// send API call to create the pod
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodInterface
	_, err := c.Kubernetes.CoreV1().
		Pods(c.config.Namespace).
		Create(context.Background(), c.Pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
