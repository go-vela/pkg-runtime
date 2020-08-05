// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"github.com/go-vela/types/constants"

	"github.com/urfave/cli/v2"
)

// Flags represents all supported command line
// interface (CLI) flags for the runtime.
//
// https://pkg.go.dev/github.com/urfave/cli?tab=doc#Flag
var Flags = []cli.Flag{

	// Logging Flags

	&cli.StringFlag{
		EnvVars: []string{"RUNTIME_LOG_LEVEL", "VELA_LOG_LEVEL", "LOG_LEVEL"},
		Name:    "runtime.log.level",
		Usage:   "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
		Value:   "info",
	},

	// Runtime Flags

	&cli.StringFlag{
		EnvVars: []string{"VELA_RUNTIME_DRIVER", "RUNTIME_DRIVER"},
		Name:    "runtime.driver",
		Usage:   "name of runtime driver to use",
		Value:   constants.DriverDocker,
	},
	&cli.StringFlag{
		EnvVars: []string{"VELA_RUNTIME_CONFIG", "RUNTIME_CONFIG"},
		Name:    "runtime.config",
		Usage:   "path to runtime configuration file",
	},
	&cli.StringFlag{
		EnvVars: []string{"VELA_RUNTIME_NAMESPACE", "RUNTIME_NAMESPACE"},
		Name:    "runtime.namespace",
		Usage:   "name of namespace for runtime configuration (kubernetes runtime only)",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"VELA_RUNTIME_VOLUMES", "RUNTIME_VOLUMES"},
		Name:    "runtime.volumes",
		Usage:   "set of volumes to mount into the runtime",
	},
}
