# Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

SHELL := /bin/bash

GOLANG_DOCKER_IMAGE := golang:1.24

IMAGE_NAME := $(shell basename "$$(pwd)")-app
BUILDER := extend-builder

build: proto
	docker run -t --rm \
			-u $$(id -u):$$(id -g) \
			-e GOCACHE=/data/.cache/go-build \
			-v $$(pwd):/data \
			-w /data \
			$(GOLANG_DOCKER_IMAGE) \
			sh -c "go build -modcacherw -v"

proto:
	docker run -t --rm -u $$(id -u):$$(id -g) \
		-v $$(pwd):/data \
		-w /data \
		--entrypoint /bin/bash \
		rvolosatovs/protoc:4.1.0 \
			proto.sh

image:
	docker buildx build -t ${IMAGE_NAME} --load .

imagex:
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${IMAGE_NAME} --platform linux/amd64 .
	docker buildx build -t ${IMAGE_NAME} --load .
	docker buildx rm --keep-state $(BUILDER)

imagex_push:
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is not set (e.g. 'v0.1.0', 'latest')"; exit 1)
	@test -n "$(REPO_URL)" || (echo "REPO_URL is not set"; exit 1)
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${REPO_URL}:${IMAGE_TAG} --platform linux/amd64 --push .
	docker buildx rm --keep-state $(BUILDER)