.DEFAULT_GOAL := help

# Settable variables ----------------------------------------------------------
# Set output prefix (local directory if not specified)
PREFIX ?= $(shell pwd)
# Set version (read from file if not specified)
VERSION ?= $(shell cat VERSION)
# Set image for Docker
IMGREPO ?= nocquidant/demo-hello
# Set target tag for Docker image
IMGTAG ?= latest
# Set Docker username for pushing image to registry
DOCKER_USER ?= unknown
# Set Docker password for pushing image to registry
DOCKER_PASS ?= unknown

# Compile time flags ----------------------------------------------------------
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X $(MOD)/env.GITCOMMIT=$(GITCOMMIT) -X $(MOD)/env.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# Other variables -------------------------------------------------------------
# Module name
MOD := github.com/nocquidant/demo-hello
DIST_DIR = $(PREFIX)/dist
PACKAGE = $(PREFIX)

# Util stuff ------------------------------------------------------------------
print-%: ; @echo $*=$($*)

# -----------------------------------------------------------------------------
run: ## Runs demo-hello
	GO111MODULE=on go run $(PACKAGE)

# -----------------------------------------------------------------------------
build: ## Builds demo-hello
	GO111MODULE=on go build -o $(DIST_DIR)/demo-hello $(PACKAGE)

# -----------------------------------------------------------------------------
test: ## Runs tests
	GO111MODULE=on go test -v -cover $(shell go list ./... | grep -v /vendor/ | grep -v mock)	

# -----------------------------------------------------------------------------
benchmark: ## Runs benchmarks
	GO111MODULE=on go test -bench $(shell go list ./... | grep -v /vendor/ | grep -v mock) 

# -----------------------------------------------------------------------------
define buildpackage
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 GO111MODULE=on go build -o $(DIST_DIR)/demo-hello-$(1)-$(2) -a $(GO_LDFLAGS_STATIC) $(PACKAGE)
endef

package-darwin: ${PACKAGE}/*.go ; $(call buildpackage,darwin,amd64)

package-windows: ${PACKAGE}/*.go ; $(call buildpackage,windows,amd64)

package-linux: ${PACKAGE}/*.go ; $(call buildpackage,linux,amd64)

package: package-linux package-windows package-darwin ## Cross compiles demo-hello

# -----------------------------------------------------------------------------  
docker-check-env: Dockerfile ; @which docker > /dev/null

docker-login: docker-check-env
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}

image: docker-check-env ## Builds Docker image
	docker build -f Dockerfile -t $(IMGREPO):git-$(GITCOMMIT) .

tag-image: docker-check-env ## Tags image with target tag $(IMGTAG) which should be the build #id
	docker tag $(IMGREPO):git-$(GITCOMMIT) $(IMGREPO):$(IMGTAG)

push-image: docker-login ## Pushes image to Docker Hub
	docker push $(IMGREPO)

# -----------------------------------------------------------------------------  
clean:
	rm -rf $(DIST_DIR)

# -----------------------------------------------------------------------------
help: ## Displays this help 
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)


.PHONY: run build test benchmark package docker-login image tag-image push-image clean help 
