
ARCH ?= $(shell go env GOARCH)

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

BIN_DIR := bin

# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/zawachte/inspektor-gadget-exporter:v0.0.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: binary

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

# Build binary
binary: fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/inspektor-gadget-exporter main.go

uninstall:
	kubectl delete -f config/inspector-gadget-exporter.yaml

deploy:
	kubectl apply -f config/inspector-gadget-exporter.yaml

# Build the docker image
docker-build: binary
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}