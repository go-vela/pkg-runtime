// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	kubernetes kubernetes.Interface

	namespace string
	pod       *v1.Pod

	// set of host volumes to mount into every container
	volumes []string
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
func New(namespace, path string, _volumes []string) (*client, error) {
	// use the current context in kubeconfig
	//
	// when kube config is provided use out of cluster config option else
	// function will build and return an InClusterConfig
	//
	// https://pkg.go.dev/k8s.io/client-go/tools/clientcmd?tab=doc#BuildConfigFromFlags
	config, err := clientcmd.BuildConfigFromFlags("", path)
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

	// create the client object
	return &client{
		namespace: namespace,
		pod: &v1.Pod{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		},
		kubernetes: _kubernetes,
		volumes:    _volumes,
	}, nil
}

// NewMock returns an Engine implementation that
// integrates with a Kubernetes runtime.
//
// This function is intended for running tests only.
func NewMock(_pod *v1.Pod) (*client, error) {
	return &client{
		namespace: "test",
		pod:       _pod,
		// https://pkg.go.dev/k8s.io/client-go/kubernetes/fake?tab=doc#NewSimpleClientset
		kubernetes: fake.NewSimpleClientset(_pod),
	}, nil
}
