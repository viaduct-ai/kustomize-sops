ARG GO_VERSION="1.14"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

# Step 1: Builder
FROM golang:$GO_VERSION as builder

ARG TARGETPLATFORM
ARG PKG_NAME=ksops

# Match Argo CD's build
ENV GO111MODULE=on \
    # Define kustomize config location
    XDG_CONFIG_HOME=$HOME/.config

# Export templated Go env variables
RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2) && \
    export GOARM=$(echo ${TARGETPLATFORM} | cut -d / -f3 | cut -c2-)

WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

ADD . .

# Perform the build and Install kustomize via Go
RUN make install
RUN make kustomize

# Step 2: Multi-architecture
FROM gcr.io/distroless/static:latest

COPY --from=builder /.config /.config
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/go.* /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/Makefile /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/scripts/ /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/exec_plugin.go /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/ksops.go /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/ksops /
COPY --from=builder /go/src/github.com/viaduct-ai/kustomize-sops/.git/ /

USER nobody

ENTRYPOINT ["/ksops"]