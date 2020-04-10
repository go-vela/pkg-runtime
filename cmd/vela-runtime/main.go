// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := cli.NewApp()

	// Package Information

	app.Name = "vela-runtime"
	app.HelpName = "vela-runtime"
	app.Usage = "Vela runtime package for integrating with different runtimes"
	app.Copyright = "Copyright (c) 2020 Target Brands, Inc. All rights reserved."
	app.Authors = []*cli.Author{
		{
			Name:  "Vela Admins",
			Email: "vela@target.com",
		},
	}

	// Package Metadata

	app.Compiled = time.Now()
	app.Action = run

	// Package Flags

	app.Flags = flags()

	// Package Start

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
