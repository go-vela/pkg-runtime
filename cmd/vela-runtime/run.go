// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"context"

	"github.com/go-vela/types/pipeline"

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

	// setup types
	ctx := context.Background()
	p := &pipeline.Build{
		ID: "go-vela-pkg-runtime-1",
		Steps: pipeline.ContainerSlice{
			{
				ID:          "step-go-vela-pkg-runtime-1-test",
				Commands:    []string{"echo ${FOO}"},
				Environment: map[string]string{"FOO": "bar"},
				Image:       "alpine:latest",
				Name:        "test",
				Number:      1,
				Pull:        true,
			},
		},
	}

	runtime, err := setupRuntime(c)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Creating runtime volume")

	err = runtime.CreateVolume(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	err = runtime.CreateNetwork(ctx, p)
	if err != nil {
		logrus.Fatal(err)
	}

	err = runtime.RunContainer(ctx, p, p.Steps[0])
	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}
