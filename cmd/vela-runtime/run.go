// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"bufio"
	"context"
	"fmt"

	"github.com/go-vela/pkg-runtime/runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	_ "github.com/joho/godotenv/autoload"
)

// run executes the package based off the configuration provided.
//
// nolint: funlen // ignore function length due to comments
func run(c *cli.Context) error {
	// set the log level for the plugin
	switch c.String("runtime.log.level") {
	case "t", "trace", "Trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	case "d", "debug", "Debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "w", "warn", "Warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "e", "error", "Error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "f", "fatal", "Fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "p", "panic", "Panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "i", "info", "Info", "INFO":
		fallthrough
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// setup the compiler
	compiler, err := setupCompiler(c)
	if err != nil {
		logrus.Fatal(err)
	}

	// setup the pipeline
	p, err := setupPipeline(c, compiler)
	if err != nil {
		return err
	}

	// setup the runtime
	r, err := runtime.New(&runtime.Setup{
		Driver:     c.String("runtime.driver"),
		ConfigFile: c.String("runtime.config"),
		Namespace:  c.String("runtime.namespace"),
	})
	if err != nil {
		logrus.Fatal(err)
	}

	// setup the context
	ctx := context.Background()

	logrus.Infof("creating network for pipeline %s", p.ID)
	err = r.CreateNetwork(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("creating volume for pipeline %s", p.ID)
	err = r.CreateVolume(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	defer func() {
		for _, step := range p.Steps {
			// TODO: remove hardcoded reference
			//
			// nolint: goconst // ignore init as constant
			if step.Name == "init" {
				continue
			}

			logrus.Infof("removing container for step %s", step.Name)
			// remove the runtime container
			err := r.RemoveContainer(ctx, step)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		logrus.Infof("removing volume for pipeline %s", p.ID)
		err = r.RemoveVolume(ctx, p)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("removing network for pipeline %s", p.ID)
		err = r.RemoveNetwork(ctx, p)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	for _, step := range p.Steps {
		// TODO: remove hardcoded reference
		if step.Name == "init" {
			continue
		}

		logrus.Infof("setting up container for step %s", step.Name)
		err = r.SetupContainer(ctx, step)
		if err != nil {
			return err
		}
	}

	for _, step := range p.Steps {
		// https://golang.org/doc/faq#closures_and_goroutines
		tmp := step

		// TODO: remove hardcoded reference
		if tmp.Name == "init" {
			continue
		}

		logrus.Infof("creating container for step %s", tmp.Name)
		err = r.RunContainer(ctx, tmp, p)
		if err != nil {
			return err
		}

		// tail the logs of the container
		go func() {
			logrus.Infof("tailing container for step %s", tmp.Name)
			rc, err := r.TailContainer(ctx, tmp)
			if err != nil {
				logrus.Fatal(err)
			}
			defer rc.Close()

			logrus.Infof("scanning container logs for step %s", tmp.Name)
			// create new scanner from the container output
			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				fmt.Println(string(scanner.Bytes()))
			}
		}()

		logrus.Infof("waiting for container for step %s", tmp.Name)
		err = r.WaitContainer(ctx, tmp)
		if err != nil {
			return err
		}

		logrus.Infof("inspecting container for step %s", tmp.Name)
		err = r.InspectContainer(ctx, tmp)
		if err != nil {
			return err
		}

		logrus.Infof("Container exited with code %d", tmp.ExitCode)
	}

	return nil
}
