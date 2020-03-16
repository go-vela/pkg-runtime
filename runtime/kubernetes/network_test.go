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

func TestKubernetes_CreateNetwork(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	tests := []struct {
		data *pipeline.Build
		want *v1.Pod
	}{
		{ // test build with steps
			data: &pipeline.Build{
				Version: "1",
				ID:      "__0",
				Steps: pipeline.ContainerSlice{
					{
						ID:        "step___0_init",
						Directory: "/home//",
						Image:     "#init",
						Name:      "init",
						Number:    1,
					},
					{
						ID:        "step___0_clone",
						Directory: "/home//",
						Image:     "target/vela-git:v0.3.0",
						Name:      "clone",
						Number:    2,
					},
					{
						ID:         "step___0_install",
						Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
						Directory:  "/home//",
						Entrypoint: []string{"/bin/sh", "-c"},
						Image:      "alpine:latest",
						Name:       "install",
						Number:     3,
					},
				},
			},
			want: r.pod,
		},
		{ // test build with steps and services
			data: &pipeline.Build{
				Version: "1",
				ID:      "__0",
				Services: pipeline.ContainerSlice{&pipeline.Container{
					ID:     "services___0_postgres",
					Image:  "postgres:latest",
					Name:   "postgres",
					Number: 1,
				}},
				Steps: pipeline.ContainerSlice{
					{
						ID:        "step___0_init",
						Directory: "/home//",
						Image:     "#init",
						Name:      "init",
						Number:    1,
					},
					{
						ID:        "step___0_clone",
						Directory: "/home//",
						Image:     "target/vela-git:v0.3.0",
						Name:      "clone",
						Number:    2,
					},
					{
						ID:         "step___0_install",
						Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
						Directory:  "/home//",
						Entrypoint: []string{"/bin/sh", "-c"},
						Image:      "alpine:latest",
						Name:       "install",
						Number:     3,
					},
				},
			},
			want: r.pod,
		},
		{ // test build with stages
			data: &pipeline.Build{Version: "1",
				ID: "__0",
				Stages: pipeline.StageSlice{
					{
						Name: "init",
						Steps: pipeline.ContainerSlice{
							{
								ID:        "__0_init_init",
								Directory: "/home//",
								Image:     "#init",
								Name:      "init",
								Number:    1,
							},
						},
					},
					{
						Name: "clone",
						Steps: pipeline.ContainerSlice{
							{
								ID:        "__0_clone_clone",
								Directory: "/home//",
								Image:     "target/vela-git:v0.3.0",
								Name:      "clone",
								Number:    2,
							},
						},
					},
					{
						Name:  "install",
						Needs: []string{"clone"},
						Steps: pipeline.ContainerSlice{
							{
								ID:         "step___0_install",
								Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
								Directory:  "/home//",
								Entrypoint: []string{"/bin/sh", "-c"},
								Image:      "alpine:latest",
								Name:       "install",
								Number:     3,
							},
						},
					},
				},
			},
			want: r.pod,
		},
	}

	// run test
	for _, test := range tests {
		// run test
		err := r.CreateNetwork(context.Background(), test.data)
		if err != nil {
			t.Errorf("CreateNetwork should not have returned err: %w", err)
		}

		if !reflect.DeepEqual(r.pod, test.want) {
			t.Errorf("Pod is %v, want %v", r.pod, test.want)
		}
	}
}

func TestKubernetes_InspectNetwork(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
		Steps: pipeline.ContainerSlice{
			{
				ID:        "step___0_init",
				Directory: "/home//",
				Image:     "#init",
				Name:      "init",
				Number:    1,
			},
			{
				ID:        "step___0_clone",
				Directory: "/home//",
				Image:     "target/vela-git:v0.3.0",
				Name:      "clone",
				Number:    2,
			},
			{
				ID:         "step___0_install",
				Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
				Directory:  "/home//",
				Entrypoint: []string{"/bin/sh", "-c"},
				Image:      "alpine:latest",
				Name:       "install",
				Number:     3,
			},
		},
	}

	_ = r.CreateNetwork(context.Background(), b)

	want, _ := json.Marshal(r.pod.Spec.HostAliases)

	// run test
	got, err := r.InspectNetwork(context.Background(), b)
	if err != nil {
		t.Errorf("InspectNetwork should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Pod is %v, want %v", got, want)
	}
}

func TestKubernetes_RemoveNetwork(t *testing.T) {
	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
		Steps: pipeline.ContainerSlice{
			{
				ID:        "step___0_init",
				Directory: "/home//",
				Image:     "#init",
				Name:      "init",
				Number:    1,
			},
			{
				ID:        "step___0_clone",
				Directory: "/home//",
				Image:     "target/vela-git:v0.3.0",
				Name:      "clone",
				Number:    2,
			},
			{
				ID:         "step___0_install",
				Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
				Directory:  "/home//",
				Entrypoint: []string{"/bin/sh", "-c"},
				Image:      "alpine:latest",
				Name:       "install",
				Number:     3,
			},
		},
	}

	_ = r.CreateNetwork(context.Background(), b)

	want := r.pod

	// run test
	err := r.RemoveNetwork(context.Background(), b)
	if err != nil {
		t.Errorf("RemoveNetwork should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(r.pod, want) {
		t.Errorf("Pod is %v, want %v", r.pod, want)
	}
}
