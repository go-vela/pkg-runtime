// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"testing"
)

func TestKubernetes_New_Success(t *testing.T) {
	// setup types
	tests := []struct {
		namespace string
		path      string
		want      client
	}{
		{
			namespace: "test",
			path:      "testdata/config",
		},
	}

	// run tests
	for _, test := range tests {
		// run test
		_, err := New(test.namespace, test.path)
		if err != nil {
			t.Errorf("New should not have returned err: %w", err)
		}
	}
}

func TestKubernetes_New_Failure(t *testing.T) {
	// setup types
	tests := []struct {
		namespace string
		path      string
		want      client
	}{
		{ // should throw error with empty config
			namespace: "test",
			path:      "testdata/config_empty",
		},
	}

	// run tests
	for _, test := range tests {
		// run test
		_, err := New(test.namespace, test.path)
		if err == nil {
			t.Error("New should have returned err")
		}
	}
}
