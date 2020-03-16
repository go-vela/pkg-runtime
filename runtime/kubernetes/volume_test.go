// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-vela/types/pipeline"
	v1 "k8s.io/api/core/v1"
)

func TestKubernetes_CreateVolume(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
	}

	want := r.pod

	want.Spec.Volumes = append(want.Spec.Volumes, v1.Volume{
		Name: b.ID,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	// run test
	err := r.CreateVolume(context.Background(), b)
	if err != nil {
		t.Errorf("CreateVolume should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(r.pod, want) {
		t.Errorf("Pod is %v, want %v", r.pod, want)
	}
}

func TestKubernetes_InspectVolume(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
	}

	_ = r.CreateVolume(context.Background(), b)

	want, _ := json.Marshal(r.pod.Spec.Volumes)

	// run test
	got, err := r.InspectVolume(context.Background(), b)
	if err != nil {
		t.Errorf("InspectVolume should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("bytes is %v, want %v", string(got), string(want))
	}
}

func TestKubernetes_RemoveVolume(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
	}

	_ = r.CreateVolume(context.Background(), b)

	want := r.pod

	// run test
	err := r.RemoveVolume(context.Background(), b)
	if err != nil {
		t.Errorf("RemoveVolume should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(r.pod, want) {
		t.Errorf("Pod is %v, want %v", r.pod, want)
	}
}
