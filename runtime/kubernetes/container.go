// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/base64"
	"io"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

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
	// TODO: do something with this
	c.Pod.ObjectMeta = metav1.ObjectMeta{Name: b.ID}

	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		return err
	}

	container := v1.Container{
		Name:      ctn.ID,
		Image:     image,
		Env:       []v1.EnvVar{},
		Stdin:     false,
		StdinOnce: false,
		TTY:       false,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      b.ID,
				MountPath: "/home",
			},
		},
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

	// // check if the entrypoint is provided
	// if len(ctn.Entrypoint) > 0 {
	// 	// add entrypoint to container config
	// 	container.Command = ctn.Entrypoint
	// }

	// // check if the commands are provided
	// if len(ctn.Commands) > 0 {
	// 	// add commands to container config
	// 	container.Command = ctn.Commands
	// }

	script := `
echo $ echo ${FOO}
echo ${FOO}
`

	baseScript := base64.StdEncoding.EncodeToString([]byte(script))

	// set the environment variables for the step
	container.Env = append(container.Env,
		v1.EnvVar{Name: "VELA_BUILD_SCRIPT", Value: baseScript},
	)
	container.Env = append(container.Env,
		v1.EnvVar{Name: "HOME", Value: "/root"},
	)
	container.Env = append(container.Env,
		v1.EnvVar{Name: "SHELL", Value: "/bin/sh"},
	)

	container.Command = []string{"/bin/sh", "-c"}
	container.Args = []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"}

	c.Pod.Spec.Containers = append(c.Pod.Spec.Containers, container)

	logrus.Infof("Creating pod %s", c.Pod.ObjectMeta.Name)
	pod, err := c.Runtime.CoreV1().Pods("docker").Create(c.Pod)
	if err != nil {
		return err
	}
	logrus.Infof("Pod created: %+v", pod)

	return nil
}

// SetupContainer pulls the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
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
