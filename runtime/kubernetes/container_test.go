// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-vela/types/pipeline"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetes_InspectContainer(t *testing.T) {
	want := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "go-vela-pkg-runtime-1",
			Namespace: "test",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "step---1-init",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
				{
					Name: "step---2-clone",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
				{
					Name: "step---3-echo",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "step---1-init",
					Image: "no-op",
				},
				{
					Name:  "step---2-clone",
					Image: "target/vela-git:latest",
				},
				{
					Name:  "step---3-echo",
					Image: "alpine:latest",
				},
			},
		},
	}

	// setup types
	c := &pipeline.Container{
		ID:         "step---3-echo",
		Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
		Directory:  "/home//",
		Entrypoint: []string{"/bin/sh", "-c"},
		Image:      "alpine:latest",
		Name:       "echo",
		Number:     3,
	}

	// setup kubernetes
	r, _ := NewMock("test", want)
	r.pod = want

	// run test
	err := r.InspectContainer(context.Background(), c)
	if err != nil {
		t.Errorf("InspectContainer should not have returned err: %w", err)
	}
}

func TestKubernetes_RemoveContainer(t *testing.T) {
	// setup types
	want := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "go-vela-pkg-runtime-1",
			Namespace: "test",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "step---1-init",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
				{
					Name: "step---2-clone",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
				{
					Name: "step---3-echo",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "Completed",
							ExitCode: 0,
						},
					},
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "step---1-init",
					Image: "no-op",
				},
				{
					Name:  "step---2-clone",
					Image: "target/vela-git:latest",
				},
				{
					Name:  "step---3-echo",
					Image: "alpine:latest",
				},
			},
		},
	}

	b := &pipeline.Container{
		ID:         "step---3-echo",
		Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
		Directory:  "/home//",
		Entrypoint: []string{"/bin/sh", "-c"},
		Image:      "alpine:latest",
		Name:       "echo",
		Number:     3,
	}

	// setup kubernetes
	r, _ := NewMock("test", want)
	r.pod = want

	// run test
	err := r.RemoveContainer(context.Background(), b)
	if err != nil {
		t.Errorf("RemoveContainer should not have returned err: %w", err)
	}
}

func TestKubernetes_RunContainer(t *testing.T) {
	// setup kubernetes
	want := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "step---1-init",
					State: v1.ContainerState{},
				},
				{
					Name:  "step---2-clone",
					State: v1.ContainerState{},
				},
				{
					Name:  "step---3-echo",
					State: v1.ContainerState{},
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "step---1-init",
					Image: "no-op",
				},
				{
					Name:  "step---2-clone",
					Image: "target/vela-git:latest",
				},
				{
					Name:  "step---3-echo",
					Image: "alpine:latest",
				},
			},
		},
	}

	r, _ := NewMock("test", want)

	// setup types
	b := &pipeline.Build{
		Version: "1",
		ID:      "__0",
	}

	r.pod = want

	// setup types
	c := &pipeline.Container{
		ID:         "step---3-echo",
		Commands:   []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
		Directory:  "/home//",
		Entrypoint: []string{"/bin/sh", "-c"},
		Image:      "alpine:latest",
		Name:       "echo",
		Number:     3,
	}

	// run test
	err := r.RunContainer(context.Background(), c, b)
	if err != nil {
		t.Errorf("RunContainer should not have returned err: %w", err)
	}
}

func TestKubernetes_SetupContainer(t *testing.T) {
	// setup types
	c := &pipeline.Container{
		ID:          "step---3-echo",
		Commands:    []string{"echo $VELA_BUILD_SCRIPT | base64 -d | /bin/sh -e"},
		Directory:   "/home//",
		Entrypoint:  []string{"/bin/sh", "-c"},
		Environment: map[string]string{"foo": "bar"},
		Image:       "alpine:latest",
		Name:        "echo",
		Number:      3,
	}

	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})
	want := r.pod

	// run test
	err := r.SetupContainer(context.Background(), c)
	if err != nil {
		t.Errorf("SetupContainer should not have returned err: %w", err)
	}

	if !reflect.DeepEqual(r.pod, want) {
		t.Errorf("Pod is %v, want %v", r.pod, want)
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
