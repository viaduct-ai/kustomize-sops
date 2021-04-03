ARG GO_VERSION="1.14-alpine"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

FROM golang:$GO_VERSION

ARG TARGETPLATFORM
ARG PKG_NAME=ksops

# Match Argo CD's build
ENV GO111MODULE=on \
    # Define kustomize config location
    XDG_CONFIG_HOME=$HOME/.config \
		CGO_ENABLED=0

# Run updates and add basic packages
RUN apk add --no-cache --update git gcc make musl-dev build-base

# Export templated Go env variables
RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2) && \
    export GOARM=$(echo ${TARGETPLATFORM} | cut -d / -f3 | cut -c2-)

WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

ADD . .

# Perform the build
RUN make install

# Install kustomize via Go
RUN make kustomize

CMD ["kustomize", "version"]
