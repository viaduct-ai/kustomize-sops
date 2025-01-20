ARG GO_VERSION="1.23.5"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS base
RUN apt update && apt install git make -y
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1@sha256:923441d7c25f1e2eb5789f82d987693c47b8ed987c4ab3b075d6ed2b5d6779a3 AS xx

# Stage 1: Build KSOPS and Kustomize
FROM --platform=${BUILDPLATFORM} base AS builder
ARG TARGETPLATFORM \
    TARGETARCH \
    PKG_NAME=ksops
COPY --link --from=xx / /

# Match Argo CD's build
ENV GO111MODULE=on \
    CGO_ENABLED=0

# Define kustomize config location
ENV HOME=/root
ENV XDG_CONFIG_HOME=$HOME/.config

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    xx-go --wrap && \
    make install && \
    xx-verify --static /go/bin/ksops && \
    xx-verify --static /go/bin/kustomize-sops
RUN make kustomize

# # Stage 2: Final image
FROM --platform=${BUILDPLATFORM} gcr.io/distroless/base AS runtime
LABEL org.opencontainers.image.source="https://github.com/viaduct-ai/kustomize-sops"

USER nonroot

WORKDIR /usr/local/bin

CMD ["kustomize", "version"]

COPY --link --from=builder --chown=root:root --chmod=755 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --link --from=base --chown=root:root --chmod=755 /usr/bin/git /usr/bin/git

# Copy only necessary files from the builder stage
COPY --link --from=builder --chown=root:root --chmod=755 /go/bin/ksops /usr/local/bin/ksops
COPY --link --from=builder --chown=root:root --chmod=755 /go/bin/kustomize /usr/local/bin/kustomize
COPY --link --from=builder --chown=root:root --chmod=755 /go/bin/kustomize-sops /usr/local/bin/kustomize-sops
