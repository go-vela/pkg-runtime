// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"testing"

	"github.com/go-vela/types/pipeline"
	v1 "k8s.io/api/core/v1"
)

func TestKubernetes_InspectImage(t *testing.T) {
	// setup types
	c := &pipeline.Container{
		ID:        "__0_init_init",
		Directory: "/home//",
		Image:     "#init",
		Name:      "init",
		Number:    1,
	}

	// setup kubernetes
	r, _ := NewMock("test", &v1.Pod{})

	// run test
	got, err := r.InspectImage(context.Background(), c)
	if err != nil {
		t.Errorf("InspectImage should not have returned err: %w", err)
	}

	if !(got == nil) {
		t.Errorf("Bytes are %v, want %v", got, nil)
	}
}
