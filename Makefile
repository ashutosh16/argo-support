SHELL := /bin/bash

VERSION := "local"
GIT_COMMIT :=  $(strip $(shell git rev-parse HEAD))

#REGISTRY := docker.intuit.com/dev-build/ibp-genai-service/service
IMAGE_NAME := ibp-genai-service
#TAG_NAME := $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

export GOPRIVATE=github.intuit.com
export GITHUB_PROJECT=dev-build/ibp-genai-service

export APP_ENV ?= dev

.PHONY: run
run: lint
	CGO_ENABLED=0 go run -ldflags "-X main.Version=$(VERSION) -X main.SHA=$(GIT_COMMIT)" ./cmd/ibp-genai-service/main.go

.PHONY: lint
lint:
	golangci-lint run --timeout 10m

.PHONY: test
test:
	go test -timeout=180s -coverprofile=coverage.out github.intuit.com/${GITHUB_PROJECT}/...
	@grep -v \
		-e "internal/genai/express" \
		-e "cmd/ibp-genai-service" \
		coverage.out > filtered_coverage.out

.PHONY: docker-build
docker-build:
	@docker build --build-arg GITHUB_INTUIT_TOKEN=${GITHUB_INTUIT_TOKEN} --build-arg VERSION=$(VERSION) --build-arg=GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local .

.PHONY: docker-test
docker-test:
	docker build --target build --build-arg GITHUB_INTUIT_TOKEN=${GITHUB_INTUIT_TOKEN} -t test_build .
	docker run test_build
