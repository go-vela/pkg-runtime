// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := cli.NewApp()

	// Package Information

	app.Name = "vela-runtime"
	app.HelpName = "vela-runtime"
	app.Usage = "Vela runtime package for integrating with different runtimes"
	app.Copyright = "Copyright (c) 2020 Target Brands, Inc. All rights reserved."
	app.Authors = []cli.Author{
		{
			Name:  "Vela Admins",
			Email: "vela@target.com",
		},
	}

	// Package Metadata

	app.Compiled = time.Now()
	app.Action = run

	// Package Flags

	app.Flags = []cli.Flag{

		cli.StringFlag{
			EnvVar: "PACKAGE_LOG_LEVEL,VELA_LOG_LEVEL,RUNTIME_LOG_LEVEL",
			Name:   "log.level",
			Usage:  "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Value:  "info",
		},

		// Runtime Flags

		cli.StringFlag{
			EnvVar: "PACKAGE_RUNTIME_DRIVER,VELA_RUNTIME_DRIVER,RUNTIME_DRIVER",
			Name:   "runtime.driver",
			Usage:  "name of runtime driver to use",
		},
		cli.StringFlag{
			EnvVar: "PACKAGE_RUNTIME_PATH,VELA_RUNTIME_PATH,RUNTIME_PATH",
			Name:   "runtime.path",
			Usage:  "path to runtime configuration file",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
