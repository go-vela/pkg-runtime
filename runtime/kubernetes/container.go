// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

const namespace = "docker"

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

	// TODO: investigate way to move this logic
	//
	// check if the pod is already created
	if len(c.Pod.ObjectMeta.Name) == 0 {
		// TODO: investigate way to make this cleaner
		//
		// iterate through each container in the pod
		for _, container := range c.Pod.Spec.Containers {
			// update the container with the volume to mount
			container.VolumeMounts = []v1.VolumeMount{
				{
					Name:      b.ID,
					MountPath: "/home",
				},
			}
		}

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

		// send API call to create the pod
		logrus.Infof("Creating pod %s", c.Pod.ObjectMeta.Name)
		_, err := c.Runtime.CoreV1().Pods(namespace).Create(c.Pod)
		if err != nil {
			return err
		}
	}

	// capture the container number for the pod
	//
	// We subtract 2 from the original step number because
	// we skip the "init" step that's automatically injected
	// into every pipeline.
	number := ctn.Number - 2

	logrus.Debugf("parsing image for container %s", ctn.Name)
	// parse image from step
	image, err := parseImage(ctn.Image)
	if err != nil {
		return err
	}

	// set the image for the container to the parsed image
	c.Pod.Spec.Containers[number].Image = image

	patch := fmt.Sprintf(patchPattern, number, c.Pod.Spec.Containers[number].Image)
	logrus.Debugf("patch: %s", patch)

	logrus.Infof("Patching pod %s", c.Pod.ObjectMeta.Name)
	// send API call to patch the pod with the new image
	_, err = c.Runtime.CoreV1().Pods(namespace).Patch(
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

	// create the container object for the pod
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#Container
	container := v1.Container{
		Name: ctn.ID,
		// create the container with the kubernetes/pause image
		//
		// This is done due to the nature of how containers are
		// executed inside the pod. Kubernetes will attempt to
		// start and run all containers in the pod at once. We
		// want to control the execution of the containers
		// inside the pod so we use the pause image as the
		// default for containers, and then sequentially patch
		// the containers with the proper image.
		//
		// https://hub.docker.com/r/kubernetes/pause
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

	// add the container definition to the pod spec
	c.Pod.Spec.Containers = append(c.Pod.Spec.Containers, container)

	return nil
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	logrus.Tracef("tailing step %s", ctn.Name)

	// create object to store container logs
	var logs io.ReadCloser

	// create function for periodically capturing
	// the logs from the container with backoff
	logsFunc := func() (bool, error) {
		// create options for capturing the logs from the container
		//
		// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#PodLogOptions
		opts := &v1.PodLogOptions{
			Container:  ctn.ID,
			Follow:     true,
			Previous:   false,
			Timestamps: false,
		}

		// send API call to capture stream of container logs
		//
		// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodExpansion
		// ->
		// https://pkg.go.dev/k8s.io/client-go/rest?tab=doc#Request.Stream
		stream, err := c.Runtime.CoreV1().
			Pods(namespace).
			GetLogs(c.Pod.ObjectMeta.Name, opts).
			Stream()
		if err != nil {
			logrus.Errorf("unable to capture logs for container %s: %v", ctn.ID, err)
			return false, nil
		}

		// create temporary reader to ensure logs are available
		reader := bufio.NewReader(stream)

		// peek at container logs from the stream
		bytes, err := reader.Peek(5)
		if err != nil {
			// skip so we resend API call to capture stream
			return false, nil
		}

		// check if we have container logs from the stream
		if len(bytes) > 0 {
			// set the logs to the reader
			logs = ioutil.NopCloser(reader)
			return true, nil
		}

		// no logs are available
		return false, nil
	}

	// create backoff object for capturing the logs
	// from the container with periodic backoff
	//
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/util/wait?tab=doc#Backoff
	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2.0,
		Jitter:   0.25,
		Steps:    10,
		Cap:      1 * time.Minute,
	}

	logrus.Tracef("capturing logs for step %s with exponential backoff", ctn.Name)
	// perform the function to capture logs with periodic backoff
	//
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/util/wait?tab=doc#ExponentialBackoff
	err := wait.ExponentialBackoff(backoff, logsFunc)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// WaitContainer blocks until the pipeline container completes.
func (c *client) WaitContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("waiting for container %s", ctn.Name)

	// create label selector for watching the pod
	selector := fmt.Sprintf("pipeline=%s", c.Pod.ObjectMeta.Name)

	// create options for watching the container
	opts := metav1.ListOptions{
		LabelSelector: selector,
		Watch:         true,
	}

	// send API call to capture channel for watching the container
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodInterface
	// ->
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/watch?tab=doc#Interface
	watch, err := c.Runtime.CoreV1().Pods(namespace).Watch(opts)
	if err != nil {
		return err
	}

	for {
		// capture new result from the channel
		//
		// https://pkg.go.dev/k8s.io/apimachinery/pkg/watch?tab=doc#Interface
		result := <-watch.ResultChan()

		// convert the object from the result to a pod
		pod, ok := result.Object.(*v1.Pod)
		if !ok {
			return fmt.Errorf("unable to watch pod %s", c.Pod.ObjectMeta.Name)
		}

		// check if the pod is in a pending state
		//
		// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#PodStatus
		if pod.Status.Phase == v1.PodPending {
			// skip pod if it's in a pending state
			continue
		}

		// iterate through each container in the pod
		for _, cst := range pod.Status.ContainerStatuses {
			// check if the container has a matching ID
			//
			// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#ContainerStatus
			if !strings.EqualFold(cst.Name, ctn.ID) {
				// skip container if it's not a matching ID
				continue
			}

			// check if the container is in a terminated state
			//
			// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#ContainerState
			if cst.State.Terminated == nil {
				// skip container if it's not in a terminated state
				break
			}

			// check if the container has a terminated state reason
			//
			// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#ContainerStateTerminated
			if len(cst.State.Terminated.Reason) > 0 {
				// break watching the container as it's complete
				return nil
			}
		}
	}
}
