# Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

SHELL := /bin/bash

GOLANG_DOCKER_IMAGE := golang:1.20

IMAGE_NAME := $(shell basename "$$(pwd)")-app
BUILDER := grpc-plugin-server-builder

PROJECT_DIR ?= $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

proto:
	rm -rfv pkg/pb/*
	mkdir -p pkg/pb
	# generate the protobuf
	docker run -t --rm -u $$(id -u):$$(id -g) -v $(PROJECT_DIR):/data/ -w /data/ rvolosatovs/protoc:4.0.0 \
			--proto_path=pkg/proto \
			--go_out=pkg/pb \
			--go_opt=paths=source_relative \
			--go-grpc_out=pkg/pb \
			--go-grpc_opt=paths=source_relative \
			pkg/proto/*.proto
	# generate the swagger.json
	docker run -t --rm -u $$(id -u):$$(id -g) -v $(PROJECT_DIR):/data/ -w /data/ rvolosatovs/protoc:4.0.0 \
			--proto_path=pkg/proto \
			--grpc-gateway_out=pkg/pb \
			--grpc-gateway_opt=logtostderr=true \
			--grpc-gateway_opt=paths=source_relative \
			--openapiv2_out=apidocs \
			--openapiv2_opt=logtostderr=true \
			pkg/proto/guildService.proto

lint: proto
	rm -f lint.err
	find -type f -iname go.mod -exec dirname {} \; | while read DIRECTORY; do \
		echo "# $$DIRECTORY"; \
		docker run -t --rm -u $$(id -u):$$(id -g) -v $$(pwd):/data/ -w /data/ -e GOCACHE=/data/.cache/go-build -e GOLANGCI_LINT_CACHE=/data/.cache/go-lint golangci/golangci-lint:v1.42.1\
				sh -c "cd $$DIRECTORY && golangci-lint -v --timeout 5m --max-same-issues 0 --max-issues-per-linter 0 --color never run || touch /data/lint.err"; \
	done
	[ ! -f lint.err ] || (rm lint.err && exit 1)

build: proto
	docker run -t --rm -u $$(id -u):$$(id -g) -v $(PROJECT_DIR):/data/ -w /data/ -e GOCACHE=/data/.cache/go-build $(GOLANG_DOCKER_IMAGE) \
		sh -c "go build"

image:
	docker buildx build -t ${IMAGE_NAME} --load .

imagex: build
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${IMAGE_NAME} --platform linux/arm64/v8,linux/amd64 .
	docker buildx build -t ${IMAGE_NAME} --load .
	docker buildx rm --keep-state $(BUILDER)

imagex_push: build
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is not set (e.g. 'v0.1.0', 'latest')"; exit 1)
	@test -n "$(REPO_URL)" || (echo "REPO_URL is not set"; exit 1)
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${REPO_URL}:${IMAGE_TAG} --platform linux/arm64/v8,linux/amd64 --push .
	docker buildx rm --keep-state $(BUILDER)

test: proto
	docker run -t --rm -u $$(id -u):$$(id -g) \
		-v $$(pwd):/data/ -w /data/ \
		-e GOCACHE=/data/.cache/go-build \
		-e BASE_PATH=/guild \
		$(GOLANG_DOCKER_IMAGE) sh -c "go test -v ./..."

test_functional_local_hosted: proto
	@test -n "$(ENV_PATH)" || (echo "ENV_PATH is not set"; exit 1)
	docker build --tag service-extension-test-functional -f test/functional/Dockerfile test/functional && \
	docker run --rm -t \
		--env-file $(ENV_PATH) \
		-e GOCACHE=/data/.cache/go-build \
		-e GOPATH=/data/.cache/mod \
		-u $$(id -u):$$(id -g) \
		-v $$(pwd):/data \
		-w /data service-extension-test-functional bash ./test/functional/test-local-hosted.sh

test_functional_accelbyte_hosted: proto
	@test -n "$(ENV_PATH)" || (echo "ENV_PATH is not set"; exit 1)
ifeq ($(shell uname), Linux)
	$(eval DARGS := -u $$(shell id -u):$$(shell id -g) --group-add $$(shell getent group docker | cut -d ':' -f 3))
endif
	docker build --tag service-extension-test-functional -f test/functional/Dockerfile test/functional && \
	docker run --rm -t \
		--env-file $(ENV_PATH) \
		-e PROJECT_DIR=$(PROJECT_DIR) \
		-e GOCACHE=/data/.cache/go-build \
		-e GOPATH=/data/.cache/mod \
		-e DOCKER_CONFIG=/tmp/.docker \
		$(DARGS) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $$(pwd):/data \
		-w /data service-extension-test-functional bash ./test/functional/test-accelbyte-hosted.sh