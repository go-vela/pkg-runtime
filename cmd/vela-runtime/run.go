// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"bufio"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/joho/godotenv/autoload"
)

// run executes the package based off the configuration provided.
func run(c *cli.Context) error {
	// set the log level for the plugin
	switch c.String("log.level") {
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
	runtime, err := setupRuntime(c)
	if err != nil {
		logrus.Fatal(err)
	}

	// setup the context
	ctx := context.Background()

	logrus.Infof("Creating runtime volume")
	err = runtime.CreateVolume(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Creating runtime network")
	err = runtime.CreateNetwork(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, step := range p.Steps {
		// TODO: remove hardcoded reference
		if step.Name == "init" {
			continue
		}

		logrus.Infof("Setting up runtime container for step %s", step.Name)
		err = runtime.SetupContainer(ctx, step)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	for _, step := range p.Steps {
		// https://golang.org/doc/faq#closures_and_goroutines
		tmp := step

		// TODO: remove hardcoded reference
		if tmp.Name == "init" {
			continue
		}

		logrus.Infof("Creating runtime container for step %s", tmp.Name)
		err = runtime.RunContainer(ctx, p, tmp)
		if err != nil {
			logrus.Fatal(err)
		}

		// tail the logs of the container
		go func() {
			logrus.Infof("Starting tail process for step %s", tmp.Name)
			rc, err := runtime.TailContainer(ctx, tmp)
			if err != nil {
				logrus.Fatal(err)
			}
			defer rc.Close()

			logrus.Infof("Scanning logs for step %s", tmp.Name)
			// create new scanner from the container output
			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				fmt.Println(string(scanner.Bytes()))
			}
		}()

		logrus.Infof("waiting for step %s", tmp.Name)
		err = runtime.WaitContainer(ctx, tmp)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	return nil
}
