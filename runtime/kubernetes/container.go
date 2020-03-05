// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

const patchPattern = `[{ "op": "replace", "path": "/spec/containers/%d/image", "value": "%s" }]`

// InspectContainer inspects the pipeline container.
func (c *client) InspectContainer(ctx context.Context, ctn *pipeline.Container) error {
	return nil
}

// RemoveContainer deletes (kill, remove) the pipeline container.
func (c *client) RemoveContainer(ctx context.Context, ctn *pipeline.Container) error {
	return nil
}

// RunContainer creates and start the pipeline container.
func (c *client) RunContainer(ctx context.Context, b *pipeline.Build, ctn *pipeline.Container) error {
	logrus.Tracef("running container %s for pipeline %s", ctn.Name, b.ID)

	number := ctn.Number - 2

	logrus.Debugf("parsing image for container %s", ctn.Name)
	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		return err
	}

	// TODO: do something with this
	if len(c.Pod.ObjectMeta.Name) == 0 {
		// TODO: do something with this
		c.Pod.ObjectMeta = metav1.ObjectMeta{
			Name:   b.ID,
			Labels: map[string]string{"pipeline": b.ID},
		}
	}

	c.Pod.Spec.Containers[number].Image = image
	c.Pod.Spec.Containers[number].VolumeMounts = []v1.VolumeMount{
		{
			Name:      b.ID,
			MountPath: "/home",
		},
	}

	// check if pod is already created
	if len(c.RawPod.ObjectMeta.UID) == 0 {
		// send API call to create the pod
		logrus.Infof("Creating pod %s", c.Pod.ObjectMeta.Name)
		c.RawPod, err = c.Runtime.CoreV1().Pods("docker").Create(c.Pod)
		if err != nil {
			return err
		}

		return nil
	}

	patch := fmt.Sprintf(patchPattern, number, c.Pod.Spec.Containers[number].Image)
	logrus.Debugf("patch: %s", patch)

	logrus.Infof("Patching pod %s", c.Pod.ObjectMeta.Name)
	// send API call to update the pod
	c.RawPod, err = c.Runtime.CoreV1().Pods("docker").Patch(
		b.ID,
		types.JSONPatchType,
		[]byte(patch),
		"",
	)
	if err != nil {
		return err
	}

	return nil
}

// SetupContainer pulls the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("setting up container %s", ctn.Name)

	logrus.Tracef("creating configuration for container %s", ctn.Name)

	// create the container with the kubernetes/pause image.
	//
	// This is done due to the nature of how Kubernetes starts
	// and executes the containers in the pod. Essentially it
	// wants to execute all containers at once, where we want
	// to periodically start containers based off the pipeline.
	container := v1.Container{
		Name:            ctn.ID,
		Image:           "docker.io/kubernetes/pause:latest",
		Env:             []v1.EnvVar{},
		Stdin:           false,
		StdinOnce:       false,
		TTY:             false,
		WorkingDir:      ctn.Directory,
		ImagePullPolicy: v1.PullAlways,
	}

	// check if the environment is provided
	if len(ctn.Environment) > 0 {
		// iterate through each element in the container environment
		for k, v := range ctn.Environment {
			// add key/value environment to container config
			container.Env = append(container.Env, v1.EnvVar{Name: k, Value: v})
		}
	}

	// check if the entrypoint is provided
	if len(ctn.Entrypoint) > 0 {
		// add entrypoint to container config
		container.Args = ctn.Entrypoint
	}

	// check if the commands are provided
	if len(ctn.Commands) > 0 {
		// add commands to container config
		container.Args = append(container.Args, ctn.Commands...)
	}

	c.Pod.Spec.RestartPolicy = v1.RestartPolicyNever
	c.Pod.Spec.Containers = append(c.Pod.Spec.Containers, container)

	return nil
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	return nil, nil
}

// WaitContainer blocks until the pipeline container completes.
func (c *client) WaitContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("waiting for container %s", ctn.Name)

	r := c.Runtime

	watcher, err := r.CoreV1().Pods("docker").Watch(metav1.ListOptions{LabelSelector: "pipeline=go-vela-pkg-runtime-1", Watch: true})
	if err != nil {
		return err
	}

	for {
		e := <-watcher.ResultChan()

		pod, ok := e.Object.(*v1.Pod)
		if !ok {
			return fmt.Errorf("unable to cast pod from watcher")
		}

		for _, cst := range pod.Status.ContainerStatuses {
			// skip container if is it not the corret ID
			if !strings.EqualFold(cst.Name, ctn.ID) {
				continue
			}

			// skip container if it is not in a terminated sate
			if cst.State.Terminated == nil {
				continue
			}

			// Container exited
			if strings.EqualFold(cst.State.Terminated.Reason, "Completed") {
				// TODO: investigate constant for container state "Completed"
				return nil
			}
		}
	}
}
