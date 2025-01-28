ARG GO_VERSION="1.23.5"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

# Stage 1: Build KSOPS and Kustomize
FROM golang:$GO_VERSION AS builder

ARG TARGETPLATFORM
ARG PKG_NAME=ksops

# Match Argo CD's build
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Define kustomize config location
ENV XDG_CONFIG_HOME=$HOME/.config

# Export templated Go env variables
RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2) && \
    export GOARM=$(echo ${TARGETPLATFORM} | cut -d / -f3 | cut -c2-)

WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

COPY . .
RUN go mod download
RUN make install
RUN make kustomize

# # Stage 2: Final image
FROM debian:bullseye-slim

LABEL org.opencontainers.image.source="https://github.com/viaduct-ai/kustomize-sops"

# ca-certs and git could be required if kustomize remote-refs are used
RUN apt update -y \
    && apt install -y git ca-certificates \
    && apt clean -y && rm -rf /var/lib/apt/lists/*

# Copy only necessary files from the builder stage
COPY --from=builder /go/bin/ksops /usr/local/bin/ksops
COPY --from=builder /go/bin/kustomize /usr/local/bin/kustomize
COPY --from=builder /go/bin/kustomize-sops /usr/local/bin/kustomize-sops

# Create a symlink from /usr/local/bin/ksops to /go/bin/ksops to preserve backwards compatibility (this will be removed in a future release)
RUN mkdir -p /go/bin
RUN ln -s /usr/local/bin/ksops /go/bin/ksops
RUN ln -s /usr/local/bin/kustomize /go/bin/kustomize
RUN ln -s /usr/local/bin/kustomize-sops /go/bin/kustomize-sops
# Set GOPATH to /go to preserve backwards compatibility (this will be removed in a future release)
ENV GOPATH=/go

# Change working directory to /usr/local/bin
WORKDIR /usr/local/bin

CMD ["kustomize", "version"]
