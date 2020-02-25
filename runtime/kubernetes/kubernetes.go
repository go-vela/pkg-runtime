// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	Runtime *kubernetes.Clientset
	Pod     *v1.Pod
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
func New(path string) (*client, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	// kubeconfig is provided use out of cluster config option
	if len(path) == 0 {
		// creates the in-cluster Kubernetes configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// creates Kubernetes client from configuration
	r, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// create the client object
	c := &client{
		Runtime: r,
		Pod:     new(v1.Pod),
	}

	return c, nil
}
