// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"github.com/go-vela/pkg-runtime/runtime"

	"github.com/urfave/cli/v2"
)

func flags() []cli.Flag {
	f := []cli.Flag{

		&cli.StringFlag{
			EnvVars: []string{"VELA_PIPELINE_CONFIG", "PIPELINE_CONFIG"},
			Name:    "pipeline.config",
			Usage:   "path to pipeline configuration file",
			Value:   "runtime/testdata/steps.yml",
		},

		// Compiler Flags

		&cli.BoolFlag{
			EnvVars: []string{"VELA_COMPILER_GITHUB", "COMPILER_GITHUB"},
			Name:    "github.driver",
			Usage:   "github compiler driver",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_COMPILER_GITHUB_URL", "COMPILER_GITHUB_URL"},
			Name:    "github.url",
			Usage:   "github url, used by compiler, for pulling registry templates",
		},
		&cli.StringFlag{
			EnvVars: []string{"VELA_COMPILER_GITHUB_TOKEN", "COMPILER_GITHUB_TOKEN"},
			Name:    "github.token",
			Usage:   "github token, used by compiler, for pulling registry templates",
		},
	}

	// Runtime Flags

	f = append(f, runtime.Flags...)

	return f
}
