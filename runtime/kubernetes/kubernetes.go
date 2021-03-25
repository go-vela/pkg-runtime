// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

type config struct {
	// specifies the config file to use for the Kubernetes client
	File string
	// specifies the namespace to use for the Kubernetes client
	Namespace string
	// specifies a list of privileged images to use for the Kubernetes client
	Images []string
	// specifies a list of host volumes to use for the Kubernetes client
	Volumes []string
}

type client struct {
	config *config
	// https://pkg.go.dev/k8s.io/client-go/kubernetes#Interface
	Kubernetes kubernetes.Interface
	// https://pkg.go.dev/k8s.io/api/core/v1#Pod
	Pod *v1.Pod
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
//
// nolint: golint // ignore returning unexported client
func New(opts ...ClientOpt) (*client, error) {
	// create new Kubernetes client
	c := new(client)

	// create new fields
	c.config = new(config)
	c.Pod = new(v1.Pod)

	// apply all provided configuration options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// use the current context in kubeconfig
	//
	// when kube config is provided use out of cluster config option else
	// function will build and return an InClusterConfig
	//
	// https://pkg.go.dev/k8s.io/client-go/tools/clientcmd?tab=doc#BuildConfigFromFlags
	config, err := clientcmd.BuildConfigFromFlags("", c.config.File)
	if err != nil {
		return nil, err
	}

	// creates Kubernetes client from configuration
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes?tab=doc#NewForConfig
	_kubernetes, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// set the Kubernetes client in the runtime client
	c.Kubernetes = _kubernetes

	return c, nil
}

// NewMock returns an Engine implementation that
// integrates with a Kubernetes runtime.
//
// This function is intended for running tests only.
//
// nolint: golint // ignore returning unexported client
func NewMock(_pod *v1.Pod, opts ...ClientOpt) (*client, error) {
	// create new Kubernetes client
	c := new(client)

	// create new fields
	c.config = new(config)
	c.Pod = new(v1.Pod)

	// set the Kubernetes namespace in the runtime client
	c.config.Namespace = "test"

	// set the Kubernetes pod in the runtime client
	c.Pod = _pod

	// apply all provided configuration options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// set the Kubernetes fake client in the runtime client
	//
	// https://pkg.go.dev/k8s.io/client-go/kubernetes/fake?tab=doc#NewSimpleClientset
	c.Kubernetes = fake.NewSimpleClientset(c.Pod)

	return c, nil
}
