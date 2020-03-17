// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package runtime

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-vela/pkg-runtime/runtime/docker"
	"github.com/go-vela/pkg-runtime/runtime/kubernetes"
	"github.com/go-vela/types/constants"
)

func TestRuntime_Docker(t *testing.T) {
	// setup types
	s := &Setup{}
	want, _ := docker.New()

	// run test
	got, err := s.Docker()
	if err != nil {
		t.Errorf("Docker should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(want)) {
		t.Errorf("Docker is %+v, want %+v", got, want)
	}
}

func TestRuntime_Kubernetes(t *testing.T) {
	// setup types
	s := &Setup{Namespace: "test", Config: "testdata/config"}
	want, _ := kubernetes.New(s.Namespace, s.Config)

	// run test
	got, err := s.Kubernetes()
	if err != nil {
		t.Errorf("Kubernetes should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(want)) {
		t.Errorf("Kubernetes is %+v, want %+v", got, want)
	}
}

func TestRuntime_Validate(t *testing.T) {
	// setup types
	tests := []struct {
		data *Setup
		want error
	}{
		{
			// test if the runtime setup is empty
			data: &Setup{},
			want: fmt.Errorf("no runtime driver provided in setup"),
		},
		{
			// test if the runtime provided is set with default value
			data: &Setup{Driver: ""},
			want: fmt.Errorf("no runtime driver provided in setup"),
		},
		{
			// test if the kubernetes runtime is provided without namespace
			data: &Setup{Driver: constants.DriverKubernetes},
			want: fmt.Errorf("no runtime driver provided in setup"),
		},
	}

	// run tests
	for _, test := range tests {
		// run test
		err := test.data.Validate()
		if err == nil {
			t.Error("Validate should have returned err")
		}
	}
}
