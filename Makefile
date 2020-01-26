PROJECTNAME = kurl

WORKDIR = /go/src/$(PROJECTNAME)
GITROOT = $(shell git rev-parse --show-toplevel)

# Default values if not already in the environment
DOCKER_REPO ?= $(PROJECTNAME)
DOCKER_TAG ?= local
SOURCE_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
SOURCE_COMMIT ?= $(shell sh -c "git log -1 --pretty=oneline" | awk '{print $$1}')

OS := $(shell uname -s)
ARCH := $(shell uname -m)

ifdef DEBUG
BUILD_ARGS =--debug --binplace
# Port forward delve on debug builds
PORT_FORWARD_ARGS = -p 2345:2345
DOCKER_TARGET =deploy-dbg
TAG_SUFFIX =-dbg
else
BUILD_ARGS =--release --binplace
DOCKER_TARGET =deploy
endif 

ifeq "$(OS)" "Darwin"
BUILD_ARGS:=$(BUILD_ARGS) --mac
endif 
ifeq "$(OS)" "Linux"
BUILD_ARGS=$(BUILD_ARGS) --linux
endif

ifdef DOCKER_USER_PATH
VOLUME_ARGS=--volume $(DOCKER_USER_PATH):/user:ro
endif

.PHONY: help
help:
	@echo 'Usage: make [target] [options]'
	@echo ''
	@echo 'Targets:'      
	@echo '  dev       builds a development docker image'
	@echo '  shell     shells into the development environment'
	@echo '  build     builds /usr/local/bin/kurl in the dev environment, plus bin/kurl for your host os'
	@echo '  deploy    builds the deployable docker image'
	@echo '  clean     cleans your localhost of development artifacts'
	@echo ''
	@echo 'Options:'
	@echo '  DEBUG=1                      build the debug binary (as opposed to release)'
	@echo '  VERBOSE=1                    increase make verbosity'
	@echo '  DOCKER_REPO=[repository]     the name of a docker repository'
	@echo '  DOCKER_TAG=[tag]             a docker image tag'
	@echo '  DOCKER_USER_PATH=[directory] a readonly directory mounted in the dev container at /user,'
	@echo '                               which enables use of a .bashrc for e.g.'
	@echo ''

ifndef VERBOSE
.SILENT:
endif

config:
ifdef VERBOSE
	echo Docker=$(DOCKER_REPO):$(DOCKER_TAG) Branch=$(SOURCE_BRANCH) Commit=$(SOURCE_COMMIT) BUILD_ARGS=$(BUILD_ARGS) PORT_FORWARD_ARGS=$(PORT_FORWARD_ARGS)
	echo ""
endif

.PHONY: dev
dev: config
	docker build \
		--force-rm \
		-f $(GITROOT)/Dockerfile \
		-t $(DOCKER_REPO):dev \
		--target dev \
		$(GITROOT)

.PHONY: shell
shell: config
	docker run \
		--rm \
		-it \
		$(PORT_FORWARD_ARGS) \
		--volume $(GITROOT):$(WORKDIR) \
		$(VOLUME_ARGS) \
		--workdir $(WORKDIR) \
		--hostname devbox \
		$(DOCKER_REPO):dev

.PHONY: build
build: dev
	mkdir -p $(GITROOT)/bin
	docker run \
		--volume $(GITROOT):$(WORKDIR) \
		--workdir $(WORKDIR) \
		--entrypoint scripts/build.sh \
		--hostname devbox \
		$(DOCKER_REPO):dev \
		$(BUILD_ARGS)

.PHONY: deploy
deploy: dev
	docker build \
		--force-rm \
		-f $(GITROOT)/Dockerfile \
		-t $(DOCKER_REPO):$(DOCKER_TAG)$(TAG_SUFFIX) \
		--build-arg AUTHOR=$(USER) \
		--build-arg BRANCH=$(SOURCE_BRANCH) \
		--build-arg COMMIT=$(SOURCE_COMMIT) \
		--target $(DOCKER_TARGET) \
		$(GITROOT) 

.PHONY: clean
clean: config
	-rm -f bin/* 
	-docker image rm -f $(DOCKER_REPO):dev 1>&2 2>/dev/null
	-docker image rm -f $(DOCKER_REPO):$(DOCKER_TAG) 1>&2 2>/dev/null
	-docker image rm -f $(DOCKER_REPO):$(DOCKER_TAG)$(TAG_SUFFIX) 1>&2 2>/dev/null
