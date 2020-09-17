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

	"github.com/go-vela/pkg-runtime/internal/image"
	vol "github.com/go-vela/pkg-runtime/internal/volume"
	"github.com/go-vela/types/constants"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// InspectContainer inspects the pipeline container.
func (c *client) InspectContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("inspecting container %s", ctn.ID)

	// create options for getting the container
	opts := metav1.GetOptions{}

	// send API call to capture the container
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodInterface
	pod, err := c.kubernetes.CoreV1().Pods(c.namespace).Get(c.pod.ObjectMeta.Name, opts)
	if err != nil {
		return err
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

		// set the step exit code
		ctn.ExitCode = int(cst.State.Terminated.ExitCode)

		break
	}

	return nil
}

// RemoveContainer deletes (kill, remove) the pipeline container.
func (c *client) RemoveContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("removing container %s", ctn.ID)

	// TODO: investigate way to move this logic
	//
	// check if the pod is already created
	if len(c.pod.ObjectMeta.Name) > 0 {
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
		opts := &metav1.DeleteOptions{
			GracePeriodSeconds: &period,
			// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1?tab=doc#DeletionPropagation
			PropagationPolicy: &policy,
		}

		logrus.Infof("removing pod %s", c.pod.ObjectMeta.Name)
		// send API call to delete the pod
		err := c.kubernetes.CoreV1().Pods(c.namespace).Delete(c.pod.ObjectMeta.Name, opts)
		if err != nil {
			return err
		}

		c.pod = &v1.Pod{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		}
	}

	return nil
}

// RunContainer creates and starts the pipeline container.
func (c *client) RunContainer(ctx context.Context, ctn *pipeline.Container, b *pipeline.Build) error {
	logrus.Tracef("running container %s", ctn.ID)

	// TODO: investigate way to move this logic
	//
	// check if the pod is already created
	if len(c.pod.ObjectMeta.Name) == 0 {
		// TODO: investigate way to make this cleaner
		//
		// iterate through each container in the pod
		for _, container := range c.pod.Spec.Containers {
			// update the container with the volume to mount
			container.VolumeMounts = []v1.VolumeMount{
				{
					Name:      b.ID,
					MountPath: "/vela",
				},
			}

			// check if other volumes were provided
			if len(c.volumes) > 0 {
				// iterate through all volumes provided
				for k, v := range c.volumes {
					// parse the volume provided
					_volume := vol.Parse(v)

					// add the volume to the container
					container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
						Name:      fmt.Sprintf("%s_%d", b.ID, k),
						MountPath: _volume.Destination,
					})
				}
			}
		}

		// create the object metadata for the pod
		//
		// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1?tab=doc#ObjectMeta
		c.pod.ObjectMeta = metav1.ObjectMeta{
			Name:   b.ID,
			Labels: map[string]string{"pipeline": b.ID},
		}

		// create the restart policy for the pod
		//
		// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#RestartPolicy
		c.pod.Spec.RestartPolicy = v1.RestartPolicyNever

		logrus.Infof("creating pod %s", c.pod.ObjectMeta.Name)
		// send API call to create the pod
		//
		// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodInterface
		_, err := c.kubernetes.CoreV1().Pods(c.namespace).Create(c.pod)
		if err != nil {
			return err
		}
	}

	// parse image from step
	_image, err := image.ParseWithError(ctn.Image)
	if err != nil {
		return err
	}

	// set the pod container image to the parsed step image
	c.pod.Spec.Containers[ctn.Number-2].Image = _image

	// handle the container pull policy
	switch ctn.Pull {
	case constants.PullAlways:
		// set the pod container pull policy to always
		c.pod.Spec.Containers[ctn.Number-2].ImagePullPolicy = v1.PullAlways
	case constants.PullNever:
		// set the pod container pull policy to never
		c.pod.Spec.Containers[ctn.Number-2].ImagePullPolicy = v1.PullNever
	case constants.PullOnStart:
		// set the pod container pull policy to always
		//
		// if the pipeline container image should be pulled on start, than
		// we force Kubernetes to pull the image on start with the always
		// pull policy for the pod container
		c.pod.Spec.Containers[ctn.Number-2].ImagePullPolicy = v1.PullAlways
	case constants.PullNotPresent:
		fallthrough
	default:
		// default the pod container pull policy to if not present
		c.pod.Spec.Containers[ctn.Number-2].ImagePullPolicy = v1.PullIfNotPresent
	}

	// send API call to patch the pod with the new container image
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1?tab=doc#PodInterface
	_, err = c.kubernetes.CoreV1().Pods(c.namespace).Patch(
		c.pod.ObjectMeta.Name,
		types.StrategicMergePatchType,
		[]byte(fmt.Sprintf(imagePatch, ctn.ID, _image)),
		"",
	)
	if err != nil {
		return err
	}

	return nil
}

// SetupContainer prepares the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("setting up for container %s", ctn.ID)

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
		Image:           image.Parse("kubernetes/pause:latest"),
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
	//
	// https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#PodSpec
	c.pod.Spec.Containers = append(c.pod.Spec.Containers, container)

	return nil
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	logrus.Tracef("tailing output for container %s", ctn.ID)

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
		stream, err := c.kubernetes.CoreV1().
			Pods(c.namespace).
			GetLogs(c.pod.ObjectMeta.Name, opts).
			Stream()
		if err != nil {
			logrus.Errorf("%v", err)
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
		Cap:      2 * time.Minute,
	}

	logrus.Tracef("capturing logs with exponential backoff for container %s", ctn.ID)
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
	logrus.Tracef("waiting for container %s", ctn.ID)

	// create label selector for watching the pod
	selector := fmt.Sprintf("pipeline=%s", c.pod.ObjectMeta.Name)

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
	watch, err := c.kubernetes.CoreV1().Pods(c.namespace).Watch(opts)
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
			return fmt.Errorf("unable to watch pod %s", c.pod.ObjectMeta.Name)
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
