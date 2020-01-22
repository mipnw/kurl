FROM golang:1.13-alpine AS dev
WORKDIR /
COPY scripts/setup.sh /setup.sh
RUN /setup.sh
RUN rm setup.sh
COPY .bashrc /root/.bashrc
ENTRYPOINT ["/bin/bash"]

# Intermediate stage: source code in development environment (not meant to be a target).
# Used to group together docker build steps that are common to debug and release (simplify the Dockerfile)
FROM dev as repository
WORKDIR /go/src/kurl
COPY go.mod go.mod
COPY src src
COPY scripts scripts

# Build stage: our image built in our dev environment (not meant to be a deployable)
FROM repository as build
RUN scripts/build.sh --release 2>&1

# Deploy stage: our built image in a thinner base image
FROM alpine:latest AS deploy
COPY --from=build /usr/local/bin/kurl /usr/local/bin/kurl
ARG AUTHOR
ARG BRANCH
ARG COMMIT
LABEL Author=$AUTHOR Branch=$BRANCH Commit=$COMMIT
ENTRYPOINT ["/usr/local/bin/kurl"]

# Build stage: our image built in our dev environment (not meant to be a deployable)
FROM repository as build-dbg
RUN scripts/build.sh --debug 2>&1

# Deploy stage: our debug built image in a thinner base image
FROM alpine:latest AS deploy-dbg
COPY --from=build-dbg /usr/local/bin/kurl /usr/local/bin/kurl
ARG AUTHOR
ARG BRANCH
ARG COMMIT
LABEL Author=$AUTHOR Branch=$BRANCH Commit=$COMMIT
