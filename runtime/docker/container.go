// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-vela/types/pipeline"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/sirupsen/logrus"
)

// InspectContainer inspects the pipeline container.
func (c *client) InspectContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("Inspecting container for step %s", ctn.ID)

	// send API call to inspect the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerInspect
	container, err := c.Runtime.ContainerInspect(ctx, ctn.ID)
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
	logrus.Tracef("Removing container for step %s", ctn.ID)

	// send API call to inspect the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerInspect
	container, err := c.Runtime.ContainerInspect(ctx, ctn.ID)
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
		err := c.Runtime.ContainerKill(ctx, ctn.ID, "SIGKILL")
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
	err = c.Runtime.ContainerRemove(ctx, ctn.ID, opts)
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

// RunContainer creates and start the pipeline container.
func (c *client) RunContainer(ctx context.Context, ctn *pipeline.Container, b *pipeline.Build) error {
	// create container configuration
	c.ctnConf = ctnConfig(ctn)
	c.netConf = netConfig(b.ID, ctn.Name)

	logrus.Tracef("Creating container for step %s", b.ID)

	// send API call to create the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerCreate
	container, err := c.Runtime.ContainerCreate(
		ctx,
		// https://godoc.org/github.com/docker/docker/api/types/container#Config
		c.ctnConf,
		// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
		c.hostConf,
		// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
		c.netConf,
		ctn.ID,
	)
	if err != nil {
		return err
	}

	logrus.Tracef("Starting container for step %s", b.ID)

	// create options for starting container
	//
	// https://godoc.org/github.com/docker/docker/api/types#ContainerStartOptions
	opts := types.ContainerStartOptions{}

	// send API call to start the container
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerStart
	err = c.Runtime.ContainerStart(ctx, container.ID, opts)
	if err != nil {
		return err
	}

	return nil
}

// SetupContainer pulls the image for the pipeline container.
func (c *client) SetupContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("Parsing image %s", ctn.Image)

	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		return err
	}

	// check if the container should be updated
	if ctn.Pull {
		logrus.Tracef("Pulling configured image %s", image)
		// create options for pulling image
		//
		// https://godoc.org/github.com/docker/docker/api/types#ImagePullOptions
		opts := types.ImagePullOptions{}

		// send API call to pull the image for the container
		//
		// https://godoc.org/github.com/docker/docker/client#Client.ImagePull
		reader, err := c.Runtime.ImagePull(ctx, image, opts)
		if err != nil {
			return err
		}

		defer reader.Close()

		// copy output from image pull to standard output
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return err
		}

		return nil
	}

	// check if the container image exists on the host
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ImageInspectWithRaw
	_, _, err = c.Runtime.ImageInspectWithRaw(ctx, image)
	if err == nil {
		return nil
	}

	// if the container image does not exist on the host
	// we attempt to capture it for executing the pipeline
	//
	// https://godoc.org/github.com/docker/docker/client#IsErrNotFound
	if docker.IsErrNotFound(err) {
		logrus.Tracef("Pulling unfound image %s", image)

		// create options for pulling image
		//
		// // https://godoc.org/github.com/docker/docker/api/types#ImagePullOptions
		opts := types.ImagePullOptions{}

		// send API call to pull the image for the container
		//
		// https://godoc.org/github.com/docker/docker/client#Client.ImagePull
		reader, err := c.Runtime.ImagePull(ctx, image, opts)
		if err != nil {
			return err
		}

		defer reader.Close()

		// copy output from image pull to standard output
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}

// TailContainer captures the logs for the pipeline container.
func (c *client) TailContainer(ctx context.Context, ctn *pipeline.Container) (io.ReadCloser, error) {
	logrus.Tracef("Capturing container logs for step %s", ctn.ID)

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
	logs, err := c.Runtime.ContainerLogs(ctx, ctn.ID, opts)
	if err != nil {
		return nil, err
	}

	// create in-memory pipe for capturing logs
	rc, wc := io.Pipe()

	logrus.Tracef("Copying container logs for step %s", ctn.ID)

	// capture all stdout and stderr logs
	go func() {
		// copy container stdout and stderr logs to our in-memory pipe
		//
		// https://godoc.org/github.com/docker/docker/pkg/stdcopy#StdCopy
		_, err := stdcopy.StdCopy(wc, wc, logs)
		if err != nil {
			logrus.Error("unable to copy logs: %w", err)
		}

		// close all buffers
		logs.Close()
		wc.Close()
		rc.Close()
	}()

	return rc, nil
}

// WaitContainer blocks until the pipeline container completes.
func (c *client) WaitContainer(ctx context.Context, ctn *pipeline.Container) error {
	logrus.Tracef("Waiting for container for step %s", ctn.ID)

	// send API call to wait for the container completion
	//
	// https://godoc.org/github.com/docker/docker/client#Client.ContainerWait
	wait, errC := c.Runtime.ContainerWait(ctx, ctn.ID, container.WaitConditionNotRunning)
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

	// parse image from container
	image, err := parseImage(ctn.Image)
	if err != nil {
		logrus.Errorf("unable to parse image: %v", err)
	}

	// create container config object
	//
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:        image,
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
