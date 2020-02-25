// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type client struct {
	Runtime *kubernetes.Clientset
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
func New() (*client, error) {
	// creates the in-cluster Kubernetes configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates Kubernetes client from configuration
	r, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// create the client object
	c := &client{
		Runtime: r,
	}

	return c, nil
}
