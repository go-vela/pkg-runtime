// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

const pattern = `[{ "op": "add", "path": "/spec/containers", "value": %s }]`

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

	// TODO: remove this probably
	var err error

	// TODO: do something with this
	if len(c.Pod.ObjectMeta.Name) == 0 {
		// TODO: do something with this
		c.Pod.ObjectMeta = metav1.ObjectMeta{Name: b.ID}
	}

	c.Pod.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      b.ID,
			MountPath: "/home",
		},
	}

	// sleep for 3 seconds
	time.Sleep(3 * time.Second)

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

	// logrus.Infof("Updating pod %s", c.Pod.ObjectMeta.Name)
	// // send API call to update the pod
	// c.RawPod, err = c.Runtime.CoreV1().Pods("docker").Update(c.Pod)
	// if err != nil {
	// 	return err
	// }
	// logrus.Infof("Pod updated: %+v", c.RawPod)

	logrus.Infof("Marshaling container %s for pod", ctn.Name)
	bytes, err := json.Marshal(c.Pod.Spec.Containers)
	if err != nil {
		return err
	}

	test := fmt.Sprintf(pattern, string(bytes))

	fmt.Println("Patch pattern: ", test)

	logrus.Infof("Patching pod %s", c.Pod.ObjectMeta.Name)
	// send API call to update the pod
	c.RawPod, err = c.Runtime.CoreV1().Pods("docker").Patch(
		b.ID,
		types.JSONPatchType,
		[]byte(test),
		"",
	)
	if err != nil {
		return err
	}
	logrus.Infof("Pod patched: %+v", c.RawPod)

	return nil
}

// SetupContainer pulls the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("setting up container %s", ctn.Name)

	logrus.Debugf("parsing image for container %s", ctn.Name)
	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		return err
	}

	logrus.Tracef("creating configuration for container %s", ctn.Name)
	container := v1.Container{
		Name:       ctn.ID,
		Image:      image,
		Env:        []v1.EnvVar{},
		Stdin:      false,
		StdinOnce:  false,
		TTY:        false,
		WorkingDir: ctn.Directory,
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
		container.Command = ctn.Entrypoint
	}

	// check if the commands are provided
	if len(ctn.Commands) > 0 {
		// add commands to container config
		container.Args = ctn.Commands
	}

	c.Pod.Spec.RestartPolicy = v1.RestartPolicyNever
	c.Pod.Spec.Containers = []v1.Container{container}
	// c.Pod.Spec.Containers = append(c.Pod.Spec.Containers, container)

	return nil
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	return nil, nil
}

// WaitContainer blocks until the pipeline container completes.
func (c *client) WaitContainer(ctx context.Context, ctn *pipeline.Container) error {
	return nil
}
