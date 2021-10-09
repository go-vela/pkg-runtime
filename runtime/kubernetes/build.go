// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// TODO: Vela admin defined worker-specific:
	//       NodeSelector, Tolerations, Affinity, AutomountServiceAccountToken

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

	c.createdPod = true

	return nil
}

// RemoveBuild deletes (kill, remove) the pipeline build metadata.
// This deletes the kubernetes pod.
func (c *client) RemoveBuild(ctx context.Context, b *pipeline.Build) error {
	logrus.Tracef("removing build %s", b.ID)

	if !c.createdPod {
		// nothing to do
		return nil
	}

	// create variables for the delete options
	//
	// This is necessary because the delete options
	// expect all values to be passed by reference.
	var (
		period = int64(0)
		policy = metav1.DeletePropagationForeground
	)

	// create options for removing the pod
	//
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1?tab=doc#DeleteOptions
	opts := metav1.DeleteOptions{
		GracePeriodSeconds: &period,
		// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1?tab=doc#DeletionPropagation
		PropagationPolicy: &policy,
	}

	logrus.Infof("removing pod %s", c.Pod.ObjectMeta.Name)
	// send API call to delete the pod
	err := c.Kubernetes.CoreV1().
		Pods(c.config.Namespace).
		Delete(context.Background(), c.Pod.ObjectMeta.Name, opts)
	if err != nil {
		return err
	}

	c.Pod = &v1.Pod{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
	}

	return nil
}
