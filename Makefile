# Copyright (c) 2020 Target Brands, Inc. All rights reserved.
#
# Use of this source code is governed by the LICENSE file in this repository.

.PHONY: build
build: binary-build

.PHONY: run
run: build kubernetes-run

.PHONY: docker-test
docker-test: build docker-run

.PHONY: kubernetes-test
kubernetes-test: build kubernetes-run

.PHONY: test
test: build docker-run kubernetes-run

#################################
######      Go clean       ######
#################################

.PHONY: clean
clean:

	@go mod tidy
	@go vet ./...
	@go fmt ./...
	@echo "I'm kind of the only name in clean energy right now"

#################################
######    Build Binary     ######
#################################

.PHONY: binary-build
binary-build:

	GOOS=darwin CGO_ENABLED=0 \
		go build \
		-o release/vela-runtime \
		github.com/go-vela/pkg-runtime/cmd/vela-runtime

########################################
#####          Docker Run          #####
########################################

.PHONY: docker-run
docker-run:

	release/vela-runtime \
		--runtime.log.level trace \
		--runtime.driver docker

############################################
#####          Kubernetes Run          #####
############################################

.PHONY: kubernetes-run
kubernetes-run:

	release/vela-runtime \
		--runtime.log.level trace \
		--runtime.driver kubernetes \
		--runtime.config ~/.kube/config \
		--runtime.namespace docker
