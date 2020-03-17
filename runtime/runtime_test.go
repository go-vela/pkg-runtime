// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"testing"

	"github.com/go-vela/types/constants"
)

func TestRuntime_New_Success(t *testing.T) {
	// setup types
	docker, _ := New(&Setup{Driver: constants.DriverDocker})
	kubernetes, _ := New(&Setup{Driver: constants.DriverKubernetes, Namespace: "docker", Config: "testdata/config"})

	tests := []struct {
		data *Setup
		want Engine
	}{
		{ // test for Docker runtimes
			data: &Setup{Driver: constants.DriverDocker},
			want: docker,
		},
		{ // test for Kubernetes runtime
			data: &Setup{Driver: constants.DriverKubernetes, Namespace: "docker", Config: "testdata/config"},
			want: kubernetes,
		},
	}

	// run tests
	for _, test := range tests {
		// run test
		_, err := New(test.data)
		if err != nil {
			t.Errorf("New should not have returned err: %w", err)
		}
	}
}

func TestRuntime_New_Failure(t *testing.T) {
	tests := []struct {
		data *Setup
		want Engine
	}{
		{ // test for invalid runtimes
			data: &Setup{Driver: "invalid"},
			want: nil,
		},
	}

	// run tests
	for _, test := range tests {
		// run test
		_, err := New(test.data)
		if err == nil {
			t.Error("New should have returned err")
		}
	}
}
