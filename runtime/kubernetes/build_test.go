// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"testing"

	"github.com/go-vela/types/pipeline"

	v1 "k8s.io/api/core/v1"
)

func TestKubernetes_SetupBuild(t *testing.T) {
	// setup types
	_engine, err := NewMock(&v1.Pod{})
	if err != nil {
		t.Errorf("unable to create runtime engine: %v", err)
	}

	// setup tests
	tests := []struct {
		failure  bool
		pipeline *pipeline.Build
	}{
		{
			failure:  false,
			pipeline: _stages,
		},
		{
			failure:  false,
			pipeline: _steps,
		},
	}

	// run tests
	for _, test := range tests {
		err = _engine.SetupBuild(context.Background(), test.pipeline)

		// this does not test the resulting pod spec (ie no tests for ObjectMeta, RestartPolicy)

		if test.failure {
			if err == nil {
				t.Errorf("SetupBuild should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("SetupBuild returned err: %v", err)
		}
	}
}

func TestKubernetes_AssembleBuild(t *testing.T) {
	// setup tests
	tests := []struct {
		failure  bool
		pipeline *pipeline.Build
		pod      *v1.Pod
	}{
		{
			failure:  false,
			pipeline: _stages,
			pod:      &v1.Pod{},
		},
		{
			failure:  false,
			pipeline: _steps,
			pod:      &v1.Pod{},
		},
		{
			failure:  true,
			pipeline: _stages,
			pod:      _pod,
		},
		{
			failure:  true,
			pipeline: _steps,
			pod:      _pod,
		},
	}

	// run tests
	for _, test := range tests {
		_engine, err := NewMock(test.pod)
		if err != nil {
			t.Errorf("unable to create runtime engine: %v", err)
		}

		err = _engine.AssembleBuild(context.Background(), test.pipeline)

		if test.failure {
			if err == nil {
				t.Errorf("AssembleBuild should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("AssembleBuild returned err: %v", err)
		}
	}
}

func TestKubernetes_RemoveBuild(t *testing.T) {
	// setup tests
	tests := []struct {
		failure    bool
		pipeline   *pipeline.Build
		pod        *v1.Pod
		createdPod bool
	}{
		{
			failure:    false,
			pipeline:   _stages,
			pod:        _pod,
			createdPod: true,
		},
		{
			failure:    false,
			pipeline:   _steps,
			pod:        _pod,
			createdPod: true,
		},
		{
			failure:    false,
			pipeline:   _stages,
			pod:        &v1.Pod{},
			createdPod: false,
		},
		{
			failure:    false,
			pipeline:   _steps,
			pod:        &v1.Pod{},
			createdPod: false,
		},
		{
			failure:    true,
			pipeline:   _stages,
			pod:        &v1.Pod{},
			createdPod: true,
		},
		{
			failure:    true,
			pipeline:   _steps,
			pod:        &v1.Pod{},
			createdPod: true,
		},
	}

	// run tests
	for _, test := range tests {
		_engine, err := NewMock(test.pod)
		if err != nil {
			t.Errorf("unable to create runtime engine: %v", err)
		}
		_engine.createdPod = test.createdPod

		err = _engine.RemoveBuild(context.Background(), test.pipeline)

		if test.failure {
			if err == nil {
				t.Errorf("RemoveBuild should have returned err")
			}

			continue
		}

		if err != nil {
			t.Errorf("RemoveBuild returned err: %v", err)
		}
	}
}
