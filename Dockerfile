ARG GO_VERSION="1.14"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

FROM golang:$GO_VERSION

# Match Argo CD's build
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

# Define kustomize config location
ENV XDG_CONFIG_HOME=$HOME/.config

ARG PKG_NAME=ksops

WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

ADD . .

RUN ls 

# Perform the build
RUN make install

# Install kustomize via Go
RUN make kustomize

CMD ["kustomize", "version"]
