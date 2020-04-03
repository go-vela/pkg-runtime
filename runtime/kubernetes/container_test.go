// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"testing"

	"github.com/go-vela/types/pipeline"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetes_InspectContainer(t *testing.T) {
	// setup types
	_engine, err := NewMock(_pod)
	if err != nil {
		t.Errorf("unable to create runtime engine: %v", err)
	}

	// setup tests
	tests := []struct {
		failure   bool
		container *pipeline.Container
	}{
		{
			failure:   false,
			container: _container,
		},
		{
			failure:   false,
			container: new(pipeline.Container),
		},
	}

	// run tests
	for _, test := range tests {
		err = _engine.InspectContainer(context.Background(), test.container)

		if test.failure {
			if err == nil {
				t.Errorf("InspectContainer should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("InspectContainer returned err: %v", err)
		}
	}
}

func TestKubernetes_RemoveContainer(t *testing.T) {
	// setup types
	_engine, err := NewMock(_pod)
	if err != nil {
		t.Errorf("unable to create runtime engine: %v", err)
	}

	// setup tests
	tests := []struct {
		failure   bool
		container *pipeline.Container
	}{
		{
			failure:   false,
			container: _container,
		},
	}

	// run tests
	for _, test := range tests {
		err = _engine.RemoveContainer(context.Background(), test.container)

		if test.failure {
			if err == nil {
				t.Errorf("RemoveContainer should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("RemoveContainer returned err: %v", err)
		}
	}
}

func TestKubernetes_RunContainer(t *testing.T) {
	// setup tests
	tests := []struct {
		failure   bool
		container *pipeline.Container
		pipeline  *pipeline.Build
		pod       *v1.Pod
	}{
		{
			failure:   false,
			container: _container,
			pipeline:  _stages,
			pod:       _pod,
		},
		{
			failure:   false,
			container: _container,
			pipeline:  _steps,
			pod:       _pod,
		},
		{
			failure:   false,
			container: _container,
			pipeline:  _steps,
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "step-github-octocat-1-clone",
							Image:           "target/vela-git:v0.3.0",
							WorkingDir:      "/home/github/octocat",
							ImagePullPolicy: v1.PullAlways,
						},
						{
							Name:            "step-github-octocat-1-echo",
							Image:           "alpine:latest",
							WorkingDir:      "/home/github/octocat",
							ImagePullPolicy: v1.PullAlways,
						},
					},
				},
			},
		},
	}

	// run tests
	for _, test := range tests {
		_engine, err := NewMock(test.pod)
		if err != nil {
			t.Errorf("unable to create runtime engine: %v", err)
		}

		err = _engine.RunContainer(context.Background(), test.container, test.pipeline)

		if test.failure {
			if err == nil {
				t.Errorf("RunContainer should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("RunContainer returned err: %v", err)
		}
	}
}

func TestKubernetes_SetupContainer(t *testing.T) {
	// setup types
	_engine, err := NewMock(_pod)
	if err != nil {
		t.Errorf("unable to create runtime engine: %v", err)
	}

	// setup tests
	tests := []struct {
		failure   bool
		container *pipeline.Container
	}{
		{
			failure:   false,
			container: _container,
		},
		{
			failure: false,
			container: &pipeline.Container{
				ID:          "step_github_octocat_1_echo",
				Commands:    []string{"echo", "hello"},
				Directory:   "/home/github/octocat",
				Environment: map[string]string{"FOO": "bar"},
				Entrypoint:  []string{"/bin/sh", "-c"},
				Image:       "alpine:latest",
				Name:        "echo",
				Number:      2,
				Pull:        true,
			},
		},
	}

	// run tests
	for _, test := range tests {
		err = _engine.SetupContainer(context.Background(), test.container)

		if test.failure {
			if err == nil {
				t.Errorf("SetupContainer should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("SetupContainer returned err: %v", err)
		}
	}
}

func TestKubernetes_TailContainer(t *testing.T) {
	// TODO: investigate this test, Kubernetes mock isn't working with custom request based options
	// Current test can not be completed due to Kubernetes crashing
	// on nil request from response

}

func TestKubernetes_WaitContainer(t *testing.T) {
	// TODO: investigate using the Kubernetes test utility with Watch() clients
	// Current test can not be completed due to Kubernetes responses hanging on
	// returning the channel of events <-watch.ResultChan()
}
