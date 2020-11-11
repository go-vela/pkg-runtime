// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-vela/types/constants"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/go-vela/pkg-runtime/internal/image"
	"github.com/go-vela/types/pipeline"

	"github.com/sirupsen/logrus"
)

// InspectContainer inspects the pipeline container.
func (c *client) InspectContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("inspecting container %s", ctn.ID)

	// send API call to inspect the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerInspect
	container, err := c.docker.ContainerInspect(ctx, ctn.ID)
	if err != nil {
		return err
	}

	// capture the container exit code
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerState
	ctn.ExitCode = container.State.ExitCode

	return nil
}

// RemoveContainer deletes (kill, remove) the pipeline container.
func (c *client) RemoveContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("removing container %s", ctn.ID)

	// send API call to inspect the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerInspect
	container, err := c.docker.ContainerInspect(ctx, ctn.ID)
	if err != nil {
		return err
	}

	// if the container is paused, restarting or running
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerState
	if container.State.Paused ||
		container.State.Restarting ||
		container.State.Running {
		// send API call to kill the container
		//
		// https://godoc.org/github.com/docker/docker/client#Client.ContainerKill
		err := c.docker.ContainerKill(ctx, ctn.ID, "SIGKILL")
		if err != nil {
			return err
		}
	}

	// create options for removing container
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerRemoveOptions
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}

	// send API call to remove the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerRemove
	err = c.docker.ContainerRemove(ctx, ctn.ID, opts)
	if err != nil {
		return err
	}

	// Empty the container config
	c.ctnConf = nil

	// Empty the host config
	c.hostConf = nil

	// Empty the host config
	c.netConf = nil

	return nil
}

// RunContainer creates and starts the pipeline container.
func (c *client) RunContainer(ctx context.Context, ctn *pipeline.Container, b *pipeline.Build) error {
	logrus.Tracef("running container %s", ctn.ID)

	// check if the container pull policy is on_start
	if strings.EqualFold(ctn.Pull, constants.PullOnStart) {
		// send API call to create the image
		err := c.CreateImage(ctx, ctn)
		if err != nil {
			return err
		}
	}

	// create container configuration
	c.ctnConf = ctnConfig(ctn)
	c.netConf = netConfig(b.ID, ctn.Name)

	// send API call to create the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerCreate
	_, err := c.docker.ContainerCreate(
		ctx,
		c.ctnConf,
		c.hostConf,
		c.netConf,
		ctn.ID,
	)
	if err != nil {
		return err
	}

	// create options for starting container
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerStartOptions
	opts := types.ContainerStartOptions{}

	// send API call to start the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerStart
	err = c.docker.ContainerStart(ctx, ctn.ID, opts)
	if err != nil {
		return err
	}

	return nil
}

// SetupContainer prepares the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("setting up for container %s", ctn.ID)

	// handle the container pull policy
	switch ctn.Pull {
	case constants.PullAlways:
		// send API call to create the image
		return c.CreateImage(ctx, ctn)
	case constants.PullNotPresent:
		// handled further down in this function
		break
	case constants.PullNever:
		fallthrough
	case constants.PullOnStart:
		fallthrough
	default:
		logrus.Tracef("skipping setup for container %s due to pull policy %s", ctn.ID, ctn.Pull)

		return nil
	}

	// parse image from container
	//
	// https://pkg.go.dev/github.com/go-vela/pkg-runtime/internal/image#ParseWithError
	_image, err := image.ParseWithError(ctn.Image)
	if err != nil {
		return err
	}

	// check if the container image exists on the host
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ImageInspectWithRaw
	_, _, err = c.docker.ImageInspectWithRaw(ctx, _image)
	if err == nil {
		return nil
	}

	// if the container image does not exist on the host
	// we attempt to capture it for executing the pipeline
	//
	// https://godoc.org/github.com/docker/docker/client#IsErrNotFound
	if docker.IsErrNotFound(err) {
		// send API call to create the image
		return c.CreateImage(ctx, ctn)
	}

	return err
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	logrus.Tracef("tailing output for container %s", ctn.ID)

	// create options for capturing container logs
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerLogsOptions
	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	}

	// send API call to capture the container logs
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerLogs
	logs, err := c.docker.ContainerLogs(ctx, ctn.ID, opts)
	if err != nil {
		return nil, err
	}

	// create in-memory pipe for capturing logs
	rc, wc := io.Pipe()

	// capture all stdout and stderr logs
	go func() {
		logrus.Tracef("copying logs for container %s", ctn.ID)

		// copy container stdout and stderr logs to our in-memory pipe
		//
		// https://godoc.org/github.com/docker/docker/pkg/stdcopy#StdCopy
		_, err := stdcopy.StdCopy(wc, wc, logs)
		if err != nil {
			logrus.Errorf("unable to copy logs for container: %v", err)
		}

		// close logs buffer
		logs.Close()

		// close in-memory pipe write closer
		wc.Close()
	}()

	return rc, nil
}

// WaitContainer blocks until the pipeline container completes.
func (c *client) WaitContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("waiting for container %s", ctn.ID)

	// send API call to wait for the container completion
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerWait
	wait, errC := c.docker.ContainerWait(ctx, ctn.ID, container.WaitConditionNotRunning)

	select {
	case <-wait:
	case err := <-errC:
		return err
	}

	return nil
}

// ctnConfig is a helper function to
// generate the container config.
func ctnConfig(ctn *pipeline.Container) *container.Config {
	logrus.Tracef("Creating container configuration for step %s", ctn.ID)

	// create container config object
	//
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:        image.Parse(ctn.Image),
		WorkingDir:   ctn.Directory,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		OpenStdin:    false,
		StdinOnce:    false,
		ArgsEscaped:  false,
	}

	// check if the environment is provided
	if len(ctn.Environment) > 0 {
		// iterate through each element in the container environment
		for k, v := range ctn.Environment {
			// add key/value environment to container config
			config.Env = append(config.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// check if the entrypoint is provided
	if len(ctn.Entrypoint) > 0 {
		// add entrypoint to container config
		config.Entrypoint = ctn.Entrypoint
	}

	// check if the commands are provided
	if len(ctn.Commands) > 0 {
		// add commands to container config
		config.Cmd = ctn.Commands
	}

	return config
}
