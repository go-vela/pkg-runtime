// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	Runtime kubernetes.Interface

	namespace string
	pod       *v1.Pod
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
func New(namespace, path string) (*client, error) {
	// use the current context in kubeconfig
	//
	// when kube config is provided use out of cluster config option else
	// function will build and return an InClusterConfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	// creates Kubernetes client from configuration
	r, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// create the client object
	return &client{
		namespace: namespace,
		pod: &v1.Pod{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		},
		Runtime: r,
	}, nil
}

// New returns an Engine implementation that
// integrates with a Kubernetes runtime.
//
// This function is intended for running tests only.
func NewMock(namespace string, objects ...runtime.Object) (*client, error) {
	return &client{
		namespace: namespace,
		pod: &v1.Pod{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		},
		Runtime: fake.NewSimpleClientset(objects...),
	}, nil
}
