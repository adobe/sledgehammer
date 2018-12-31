# Copyright 2018 Adobe
# All Rights Reserved.

# NOTICE: Adobe permits you to use, modify, and distribute this file in
# accordance with the terms of the Adobe license agreement accompanying
# it. If you have received this file from a source other than Adobe,
# then your use, modification, or distribution of it requires the prior
# written permission of Adobe. 

FLAVOURS = "darwin-amd64;linux-amd64;windows-amd64"
VERSION := $(shell git describe --tags || echo "SNAPSHOT")
DATE := $(shell date)
SHA := $(shell git rev-parse HEAD)
.DEFAULT_GOAL := pr

.PHONY: pr
pr: check_requirements
	go test ./...

.PHONY: ci
ci:
	docker build -t adobe/slh:latest -t "adobe/slh:${VERSION}" --build-arg FLAVOURS=${FLAVOURS} --build-arg VERSION=${VERSION} --build-arg DATE="${DATE}" --build-arg SHA=${SHA} .

.PHONY: check_requirements
check_requirements:
	@command -v docker >/dev/null 2>&1 || { echo >&2 "docker is required but not installed. Aborting."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo >&2 "go is required but not installed. Aborting."; exit 1; }