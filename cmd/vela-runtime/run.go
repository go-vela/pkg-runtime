// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"context"
	"flag"
	"os"

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
	path := os.Getenv("KUBECONFIG")
	set := flag.NewFlagSet("test", 0)
	set.String("runtime-driver", constants.DriverKubernetes, "doc")
	set.String("runtime-path", path, "doc")
	c := cli.NewContext(nil, set, nil)
	ctx := context.Background()
	pipeline := &pipeline.Build{}

	runtime, err := setupRuntime(c)
	if err != nil {
		logrus.Fatal(err)
	}

	err = runtime.CreateVolume(ctx, pipeline)
	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}
