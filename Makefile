PROJECTNAME := kurl

WORKDIR := /go/src/$(PROJECTNAME)
GITROOT := $(shell git rev-parse --show-toplevel)

.PHONY: help
help:
	@echo 'Usage: make [target] [options]'
	@echo ''
	@echo 'Targets:'      
	@echo '  dev                 builds a development docker image'
	@echo '  shell               shells into the development environment'
	@echo '  build               runs the release build inside the development environment'
	@echo '  build-dbg           runs the debug build inside the development environment'
	@echo '  deploy              builds the deployable docker image'
	@echo '  deploy-dbg          builds the debuggable deployable docker image'
	@echo '  clean               cleans your localhost of development artifacts'
	@echo ''
	@echo 'Options:'
	@echo '  VERBOSE=1                    increase verbosity'
	@echo '  DOCKER_REPO=[repository]     the name of a docker repository'
	@echo '  DOCKER_TAG=[tag]              a docker image tag'
	@echo '  DOCKER_USER_PATH=[directory] a readonly directory mounted in the dev container at /user,'
	@echo '                               which enables use of a .bashrc for e.g.'
	@echo ''

ifndef VERBOSE
.SILENT:
endif

config:
ifndef DOCKER_REPO
	$(eval DOCKER_REPO=$(PROJECTNAME))
endif

ifndef DOCKER_TAG
	$(eval DOCKER_TAG=local)
endif

ifdef DOCKER_USER_PATH
	$(eval USER_VOLUME=--volume $(DOCKER_USER_PATH):/user:ro)
endif

ifndef SOURCE_BRANCH
	$(eval SOURCE_BRANCH=$(shell git rev-parse --abbrev-ref HEAD))
endif

ifndef SOURCE_COMMIT
	$(eval SOURCE_COMMIT=$(shell sh -c "git log -1 --pretty=oneline" | awk '{print $$1}'))
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
shell: dev
	docker run \
		--rm \
		-it \
		-p 2345:2345 \
		--volume $(GITROOT):$(WORKDIR) \
		$(USER_VOLUME) \
		--workdir $(WORKDIR) \
		--hostname devbox \
		$(DOCKER_REPO):dev

.PHONY: build
build: dev
	docker run \
		--rm \
		--volume $(GITROOT):$(WORKDIR) \
		--workdir $(WORKDIR) \
		--entrypoint scripts/build.sh \
		--hostname devbox \
		$(DOCKER_REPO):dev \
		--release --binplace

.PHONY: build
build-dbg: dev
	docker run \
		--rm \
		--volume $(GITROOT):$(WORKDIR):ro \
		--workdir $(WORKDIR) \
		--entrypoint scripts/build.sh \
		--hostname devbox \
		$(DOCKER_REPO):dev \
		--debug --binplace

.PHONY: deploy
deploy: dev
	docker build \
		--force-rm \
		-f $(GITROOT)/Dockerfile \
		-t $(DOCKER_REPO):$(DOCKER_TAG) \
		--build-arg AUTHOR=$(USER) \
		--build-arg BRANCH=$(SOURCE_BRANCH) \
		--build-arg COMMIT=$(SOURCE_COMMIT) \
		--target deploy \
		$(GITROOT) 

.PHONY: deploy
deploy-dbg: dev
	docker build \
		--force-rm \
		-f $(GITROOT)/Dockerfile \
		-t $(DOCKER_REPO):$(DOCKER_TAG)-dbg \
		--build-arg AUTHOR=$(USER) \
		--build-arg BRANCH=$(SOURCE_BRANCH) \
		--build-arg COMMIT=$(SOURCE_COMMIT) \
		--target deploy-dbg \
		$(GITROOT) 

.PHONY: clean
clean: config
	-rm -f bin/* 
	-docker image rm -f $(DOCKER_REPO):dev 1>&2 2>/dev/null
	-docker image rm -f $(DOCKER_REPO):$(DOCKER_TAG) 1>&2 2>/dev/null
	-docker image rm -f $(DOCKER_REPO):$(DOCKER_TAG)-dbg 1>&2 2>/dev/null