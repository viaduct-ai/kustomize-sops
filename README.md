# KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

 - [Overview](#overview)
 - [Requirements](#requirements)
 - [Example](#example)
 - [Development and Testing](#development-and-testing)
 - [Argo CD Integration](#argo-cd-integration)


## Overview

`KSOPS`, or kustomize-SOPS, is a [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin for SOPS encrypted resources. `KSOPS` can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin, `KSOPS` allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.

## Requirements
- [Go](https://github.com/golang/go)
- [kustomize](https://github.com/kubernetes-sigs/kustomize/) built with Go (See [details below](#kustomize-go-plugin-caveats))
- [SOPS](https://github.com/mozilla/sops)

## Example

### Download KSOPS

```bash
# export GO111MODULE=on 
go get -u github.com/viaduct-ai/kustomize-sops
```

### Build KSOPS

```bash
cd $GOPATH/src/github.com/viaduct-ai/kustomize-sops
go build -buildmode plugin -o ksops.so
```

### Build kustomize with the Same Version (v3.1.0) Used in KSOPS

```bash
# KSOPS is built with kustomize@v3.1.0 
# If you want to change versions, make sure to check that the KSOPS tests still pass 

# Remove existing kustomize executable
# rm $(which kustomize)
go get sigs.k8s.io/kustomize/v3/cmd/kustomize@v3.1.0
```

### Make the KSOPS plugin available to kustomize 

```bash
# Don't forget to define XDG_CONFIG_HOME in your .bashrc/.zshrc
# export XDG_CONFIG_HOME=$HOME/.config
mkdir -p $XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/
cp ksops.so $XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/
```

### Configure SOPS via .sops.yaml
For this example and testing, `KSOPS` relies on the `SOPS` creation rules defined in `.sops.yaml`. To make encrypted secrets more readable, we suggest using the following encryption regex to only encrypt `data` and `stringData` values. This leaves non-sensitive fields, like the secret's name, unencrypted and human readable.

```yaml
creation_rules:
  - encrypted_regex: '^(data|stringData)$'
    # Specify kms/pgp/etc encryption key
    kms: XXXXXX
```


See [SOPS](https://github.com/mozilla/sops) for details.

### Create Resources

```bash
# Create a local Kubernetes Secret
cat <<EOF > secret.yaml
kind: apiVersion: v1
Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
EOF
```

### Encrypt Resources

```bash
# Encrypt with SOPS CLI
# Specify SOPS configuration in .sops.yaml
sops -e secret.yaml > secret.enc.yaml
```

### Define KSOPS kustomize Generator
```bash
# Create a local Kubernetes Secret
cat <<EOF > secret-generator.yaml
apiVersion: viaduct.ai/v1
kind: ksops
metadata:
  # Specify a name
    name: example-secret-generator
    files:
      - ./secret.enc.yaml
EOF
```

### Create the kustomization.yaml
[Read about kustomize plugins](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/README.md)

```bash
cat <<EOF > kustomization.yaml
generators: 
  - ./secret-generator.yaml
EOF
```

### Build with kustomize 

```bash
# Build with kustomize to verify
kustomize --enable_alpha_plugins . 
```

### Troubleshooting

#### kustomize Go Plugin Caveats 
[Detailed example of kustomize Go plugin](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/goPluginGuidedExample.md)

- Validate `ksops.so` is in the `kustomize` plugin path 
    - `$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/ksops.so`
- Check your `kustomize` executable was built by Go 
    - `which kustomize`
    - `kustomize version`
- Check the Go version in `go.mod` matches the Go version used to build `kustomize`
- Check the `kustomize` version specified in `go.mod` matches the installed version of `kustomize`
    - `kustomize version`

## Development and Testing

Before developing or testing `KSOPS`, ensure all external [requirements](#requirements) are properly installed.

### Development
 
`KSOPS` implements the [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin API in `ksops.go`. 


`KSOPS`'s logic is intentionally simple. Given a list of SOPS encrypted Kubernetes manifests, it iterates over each file and decrypts it via SOPS [decrypt](https://godoc.org/go.mozilla.org/sops/decrypt) library. `KSOPS` assumes nothing about the structure of the encrypted resource and relies on [kustomize](https://github.com/kubernetes-sigs/kustomize/) for manifest validation. `KSOPS` expects the encryption key to be accessible. This is important to consider when using `KSOPS` for CI/CD.

### Testing

Testing `KSOPS` requires:

1. Configuring a encryption key and other `SOPS` configuration in `.sops.yaml`
2. Building `KSOPS` as a Go plugin
3. Copying the Go plugin to the [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin path
4. Generating encrypted test files
5. Running the Go tests

Everything but setting up `.sops.yaml` is handle for you by `scripts/run-tests.sh`. After defining `.sops.yaml`, test `KSOPS` running the following command from the repo's root directory:

```bash
# Must be run from the repo's root directory!
./scripts/run-tests.sh
```

## Argo CD Integration
 
`KSOPS` becomes even more powerful when integrated with a CI/CD pipeline. By combining `KSOPS` with [Argo CD](https://github.com/argoproj/argo-cd/), you can manage Kubernetes secrets via the same Git Ops pattern you use to manage the rest of your kubernetes manifests. To integrate `KSOPS` and [Argo CD](https://github.com/argoproj/argo-cd/), you will need to create a [custom Argo CD build](https://argoproj.github.io/argo-cd/operator-manual/custom_tools/#byoi-build-your-own-image). 

### Custom Argo CD w/ KSOPS Dockerfile
 
 ```Dockerfile
ARG ARGO_CD_VERSION="v1.2.5"
# Always match Argo CD Dockerfile's Go version!
# https://github.com/argoproj/argo-cd/blob/master/Dockerfile
ARG GO_VERSION="1.12.6"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

FROM golang:$GO_VERSION as ksops-builder

# Match Argo CD's build
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

ARG PKG_NAME=ksops

WORKDIR /go/src/github.com/viaduct-ai/

RUN git clone https://github.com/viaduct-ai/kustomize-sops.git

# CD into the clone repo
WORKDIR /go/src/github.com/viaduct-ai/kustomize-sops

# Perform the build
RUN go install
RUN go build -buildmode plugin -o ${PKG_NAME}.so ${PKG_NAME}.go

# Install kustomize via Go
RUN go install sigs.k8s.io/kustomize/v3/cmd/kustomize

#--------------------------------------------#
#--------Build Custom Argo Image-------------#
#--------------------------------------------#

FROM argoproj/argocd:$ARGO_CD_VERSION

# Switch to root for the ability to perform install
USER root

# Set the kustomize home directory
ENV XDG_CONFIG_HOME=$HOME/.config
ENV KUSTOMIZE_PLUGIN_PATH=$XDG_CONFIG_HOME/kustomize/plugin/

ARG PKG_NAME=ksops

# Override the default kustomize executable with the Go built version
COPY --from=ksops-builder /go/bin/kustomize /usr/local/bin/kustomize

# Copy the plugin to kustomize plugin path
COPY --from=ksops-builder /go/src/github.com/viaduct-ai/kustomize-sops/*  $KUSTOMIZE_PLUGIN_PATH/viaduct.ai/v1/${PKG_NAME}/

# Switch back to non-root user
USER argocd
```

Don't forget to inject any necessary credentials (i.e AWS credentials) when deploying this [Argo CD](https://github.com/argoproj/argo-cd/) + `KSOPS` build!

