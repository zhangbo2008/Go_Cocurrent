
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Container registries.
REGISTRY ?= ghcr.io/kleveross

# Container registry for base images.
BASE_REGISTRY ?= docker.io

# Image URL to use all building/pushing image targets
IMG ?= kleveross/modeljob-operator:latest

ORMB_VERSION ?= $(strip 0.0.8)
#
# These variables should not need tweaking.
#

# It's necessary to set this because some environments don't link sh -> bash.
export SHELL := /bin/bash

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

# This repo's root import path (under GOPATH).
ROOT := github.com/kleveross/klever-model-registry

# Target binaries. You can build multiple binaries for a single project.
TARGETS := model-registry modeljob-operator

CAFFE_IMAGE := nvcaffe
CAFFE_TAG := cpu-0.16.5
EXTRACT_TARGETS := caffemodel netdef graphdef h5 mxnetparams onnx pmml savedmodel torchscript
EXTRACT_IMAGE_PREFIX ?= $(strip )
EXTRACT_IMAGE_SUFFIX ?= $(strip -extract)

CONVERT_TARGETS := caffemodel_to_netdef h5_to_savedmodel mxnetparams_to_onnx netdef_to_onnx
CONVERT_IMAGE_PREFIX ?= $(strip )
CONVERT_IMAGE_SUFFIX ?= $(strip )

SERVING_TARGETS := mlserver openscoring tritonserver
SERVING_IMAGE_PREFIX ?= $(strip )
SERVING_IMAGE_SUFFIX ?= $(strip )

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Current version of the project.
VERSION ?= $(shell git describe --tags --always --dirty)
RELEASE_VERSION ?= $(VERSION)
GITSHA ?= $(shell git rev-parse --short HEAD)

# Available cpus for compiling, please refer to https://github.com/caicloud/engineering/issues/8186#issuecomment-518656946 for more information.
CPUS ?= $(shell /bin/bash hack/read_cpus_available.sh)

# Default golang flags used in build and test
# -mod=vendor: force go to use the vendor files instead of using the `$GOPATH/pkg/mod`
# -p: the number of programs that can be run in parallel
# -count: run each test and benchmark 1 times. Set this flag to disable test cache
export GOFLAGS ?= -mod=vendor -p=$(CPUS) -count=1

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build: build-local

# Run tests
test: generate fmt vet manifests
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

build-local:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED="0" go build -i -v -o $(OUTPUT_DIR)/$${target}                       \
	  -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                        \
	  	-X $(ROOT)/pkg/version.COMMIT=$(GITSHA)                                        \
	    -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                                       \
	  $(CMD_DIR)/$${target};                                                           \
	done

build-linux:
	@docker run --rm                                                                   \
	  -v $(PWD):/go/src/$(ROOT)                                                        \
	  -w /go/src/$(ROOT)                                                               \
	  -e GOOS=linux                                                                    \
	  -e GOARCH=amd64                                                                  \
	  -e GOPATH=/go                                                                    \
	  -e SHELLOPTS=$(SHELLOPTS)                                                        \
	  -e CGO_ENABLED="0"                                                               \
	  -e GO111MODULE=on                                                                \
	  -e GOFLAGS="$(GOFLAGS)"                                                          \
	  $(BASE_REGISTRY)/golang:1.13.9                                                   \
	    /bin/bash -c 'for target in $(TARGETS); do                                     \
	      go build -i -v -o $(OUTPUT_DIR)/$${target}                                   \
	        -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                  \
			  -X $(ROOT)/pkg/version.COMMIT=$(GITSHA)                                  \
	          -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                                 \
	        $(CMD_DIR)/$${target};                                                     \
	    done'

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build manifests/modeljob-operator/crds | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build manifests/modeljob-operator/crds | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image modeljob-operator=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=kleveross-role webhook paths="./..." output:crd:artifacts:config=crds

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: build-linux
	@for target in $(TARGETS); do  \
		image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/$${target}/Dockerfile .;  \
	done

	# build nvcaffe: cpu

	docker build -t $(REGISTRY)/${CAFFE_IMAGE}:$(CAFFE_TAG) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/nvcaffe-cpu/Dockerfile  .; 

	# build extractor
	@for target in $(EXTRACT_TARGETS); do  \
		image=$(EXTRACT_IMAGE_PREFIX)$${target}$(EXTRACT_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/extract/$${target}/Dockerfile --build-arg ORMB_VERSION=${ORMB_VERSION} .;  \
	done

	# build convertor
	@for target in $(CONVERT_TARGETS); do  \
		image=$(CONVERT_IMAGE_PREFIX)$${target}$(CONVERT_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/convert/$${target}/Dockerfile --build-arg ORMB_VERSION=${ORMB_VERSION} .;  \
	done

	# build serving
	@for target in $(SERVING_TARGETS); do  \
		image=$(SERVING_IMAGE_PREFIX)$${target}$(SERVING_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/serving/$${target}/Dockerfile .;  \
	done

# Push the docker image
docker-push:
	@for target in $(TARGETS); do  \
		image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);   \
		docker push  $(REGISTRY)/$${image}:$(VERSION);  \
	done

	# push nvcaffe:cpu-0.16.5
	docker push  $(REGISTRY)/${CAFFE_IMAGE}:$(CAFFE_TAG); 

	# push extractor
	@for target in $(EXTRACT_TARGETS); do  \
		image=$(EXTRACT_IMAGE_PREFIX)$${target}$(EXTRACT_IMAGE_SUFFIX);   \
		docker push  $(REGISTRY)/$${image}:$(VERSION);  \
	done

	# push convertor
	@for target in $(CONVERT_TARGETS); do  \
		image=$(CONVERT_IMAGE_PREFIX)$${target}$(CONVERT_IMAGE_SUFFIX);   \
		docker push  $(REGISTRY)/$${image}:$(VERSION);  \
	done

	#push serving 
	@for target in $(SERVING_TARGETS); do  \
		image=$(SERVING_IMAGE_PREFIX)$${target}$(SERVING_IMAGE_SUFFIX);   \
		docker push  $(REGISTRY)/$${image}:$(VERSION);  \
	done

klever-docker-build-push: build
	@for target in $(TARGETS); do  \
		image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(RELEASE_VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/$${target}/Dockerfile .;  \
		docker push  $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done

	# build && push nvcaffe: cpu
	docker build -t $(REGISTRY)/${CAFFE_IMAGE}:$(CAFFE_TAG) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/nvcaffe-cpu/Dockerfile  .; 
	docker push  $(REGISTRY)/${CAFFE_IMAGE}:$(CAFFE_TAG); 
	docker rmi -f $(REGISTRY)/${CAFFE_IMAGE}:$(CAFFE_TAG); 

	# build && push extractor
	@for target in $(EXTRACT_TARGETS); do  \
		image=$(EXTRACT_IMAGE_PREFIX)$${target}$(EXTRACT_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(RELEASE_VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/extract/$${target}/Dockerfile --build-arg ORMB_VERSION=${ORMB_VERSION} .;  \
		docker push  $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done

	# build && push convertor
	@for target in $(CONVERT_TARGETS); do  \
		image=$(CONVERT_IMAGE_PREFIX)$${target}$(CONVERT_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(RELEASE_VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/convert/$${target}/Dockerfile --build-arg ORMB_VERSION=${ORMB_VERSION} .;  \
		docker push  $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done

	# build && push serving
	@for target in $(SERVING_TARGETS); do  \
		image=$(SERVING_IMAGE_PREFIX)$${target}$(SERVING_IMAGE_SUFFIX);   \
		docker build -t $(REGISTRY)/$${image}:$(RELEASE_VERSION) --label $(DOCKER_LABELS)  -f $(BUILD_DIR)/serving/$${target}/Dockerfile .;  \
		docker push  $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done
download_model:
	wget -O ormb-${ORMB_VERSION}.zip https://codeload.github.com/kleveross/ormb/zip/v${ORMB_VERSION}
	unzip -o ormb-${ORMB_VERSION}.zip  -d /tmp/

klever-extract-convert-test: download_model
	# build && push extractor
	@for target in $(EXTRACT_TARGETS); do  \
		image=$(EXTRACT_IMAGE_PREFIX)$${target}$(EXTRACT_IMAGE_SUFFIX);   \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "test $(REGISTRY)/$${image}:$(RELEASE_VERSION)"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		docker run --rm -v /tmp/ormb-${ORMB_VERSION}/examples:/models --entrypoint /bin/bash $(REGISTRY)/$${image}:$(RELEASE_VERSION) /scripts/test.sh;  \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done

	# build && push convertor
	@for target in $(CONVERT_TARGETS); do  \
		image=$(CONVERT_IMAGE_PREFIX)$${target}$(CONVERT_IMAGE_SUFFIX);   \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "test $(REGISTRY)/$${image}:$(RELEASE_VERSION)"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		echo "**********************************************"; \
		docker run --rm -v /tmp/ormb-${ORMB_VERSION}/examples:/models --entrypoint /bin/bash $(REGISTRY)/$${image}:$(RELEASE_VERSION) /scripts/test.sh;  \
		docker rmi -f $(REGISTRY)/$${image}:$(RELEASE_VERSION); \
	done	


# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

apidoc:
	@nirvana api pkg/registry --output docs/api
