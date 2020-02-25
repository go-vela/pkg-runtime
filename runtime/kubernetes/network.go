// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/go-vela/types/pipeline"
)

// CreateNetwork creates the pipeline network.
func (c *client) CreateNetwork(ctx context.Context, b *pipeline.Build) error {
	network := v1.HostAlias{
		IP:        "127.0.0.1",
		Hostnames: []string{},
	}

	for _, step := range b.Steps {
		host := fmt.Sprintf("%s.local", step.Name)

		network.Hostnames = append(network.Hostnames, host)
	}

	c.Pod.Spec.HostAliases = append(c.Pod.Spec.HostAliases, network)

	return nil
}

// InspectNetwork inspects the pipeline network.
func (c *client) InspectNetwork(ctx context.Context, b *pipeline.Build) ([]byte, error) {
	bytes, err := json.Marshal(c.Pod.Spec.HostAliases)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// RemoveNetwork deletes the pipeline network.
//
// TODO: research this
//
// currently this is a no-op because in Kubernetes the
// network lives and dies with the pod it's attached to
func (c *client) RemoveNetwork(ctx context.Context, b *pipeline.Build) error {
	return nil
}
