// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package docker

import (
	"testing"
)

func TestDocker_New(t *testing.T) {
	// setup types

	// run test
	_, err := New()
	if err != nil {
		t.Errorf("New should not have returned err: %w", err)
	}
}
