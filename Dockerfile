ARG GO_VERSION="1.14"
ARG KSOPS_REVISION="master"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

FROM golang:$GO_VERSION

# Match Argo CD's build
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

ARG PKG_NAME=ksops

WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

ADD . .

RUN git checkout $KSOPS_REVISION

# Perform the build
RUN make install

# Install kustomize via Go
RUN make kustomize

CMD ["kustomize", "version"]
