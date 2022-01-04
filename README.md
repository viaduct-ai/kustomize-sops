# KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

![Tests and Build](https://github.com/viaduct-ai/kustomize-sops/workflows/Run%20Tests%20and%20Build/badge.svg?branch=master)

- [Background](#background)
- [Overview](#overview)
- [Requirements](#requirements)
- [Installation Options](#installation-options)
- [Getting Started](#getting-started)
- [Generator Options](#generator-options)
- [Development and Testing](#development-and-testing)
- [Argo CD Integration 🤖](#argo-cd-integration-)

## Background

At [Viaduct](https://www.viaduct.ai/), we manage our Kubernetes resources via the [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) pattern; however, we could not find a solution compatible with our stack for managing secrets via the GitOps paradigm. We built `KSOPS` to connect [kustomize](https://github.com/kubernetes-sigs/kustomize/) to [SOPS](https://github.com/mozilla/sops) and integrated it with [Argo CD](https://github.com/argoproj/argo-cd) to safely manage our secrets the same way we manage the rest our Kubernetes manifest.

## Overview

`KSOPS`, or kustomize-SOPS, is a [kustomize](https://github.com/kubernetes-sigs/kustomize/) [exec plugin](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_plugins) for SOPS encrypted resources. `KSOPS` can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin, `KSOPS` allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.

## Requirements

- [kustomize](https://github.com/kubernetes-sigs/kustomize/)
- `XDG_CONFIG_HOME` environment variable is set in your shell. If it's not set, run the following

```bash
# Don't forget to define XDG_CONFIG_HOME in your .bashrc or .zshrc
echo "export XDG_CONFIG_HOME=\$HOME/.config" >> $HOME/.zshrc
source $HOME/.zshrc
```

## Installation

### Install the Latest Release

```bash
# Verify the $XDG_CONFIG_HOME environment variable exists then run
source <(curl -s https://raw.githubusercontent.com/viaduct-ai/kustomize-sops/master/scripts/install-ksops-archive.sh)
```

### Install from Source

```bash
# Optionally, install kustomize via
# make kustomize
# Verify the $XDG_CONFIG_HOME environment variable exists then run
make install
```

## Getting Started (Tutorial)

### 0. Verify Requirements

Before continuing, verify your installation of [kustomize](https://github.com/kubernetes-sigs/kustomize/)
and `gpg`. Below are a few non-comprehensive commands to quickly check your installations:

```bash
# Verify kustomize is installed
kustomize version

# Verify gpg is installed
gpg --help

# Verify XDG_CONFIG_HOME environment variable is set
echo $XDG_CONFIG_HOME
```

### 1. Download and install KSOPS

```bash
# Verify the $XDG_CONFIG_HOME environment variable exists then run
source <(curl -s https://raw.githubusercontent.com/viaduct-ai/kustomize-sops/master/scripts/install-ksops-archive.sh)
```

### 2. Import Test PGP Keys

To simplify local development and testing, we use PGP test keys. To import the keys, run the following command from the repository's root directory:

```bash
make import-test-keys
```

If you are following this tutorial, be sure to run this before the following steps. The PGP keys will also be imported when you run `make test`

See [SOPS](https://github.com/mozilla/sops) for details.

### 3. Configure SOPS via .sops.yaml

For this example and testing, `KSOPS` relies on the `SOPS` creation rules defined in `.sops.yaml`. To make encrypted secrets more readable, we suggest using the following encryption regex to only encrypt `data` and `stringData` values. This leaves non-sensitive fields, like the secret's name, unencrypted and human readable.

**Note:** You only have to modify `.sops.yaml` if you want to use your key management service in this example instead of the default PGP key imported in the previous step.

```yaml
creation_rules:
  - encrypted_regex: "^(data|stringData)$"
    # Specify kms/pgp/etc encryption key
    # This tutorial uses a local PGP key for encryption.
    # DO NOT USE IN PRODUCTION ENV
    pgp: "FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
    # Optionally you can configure to use a providers key store
    # kms: XXXXXX
    # gcp_kms: XXXXXX
```

### 4. Create a Resource

```bash
# Create a local Kubernetes Secret
cat <<EOF > secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
EOF
```

### 5. Encrypt the Resources

```bash
# Encrypt with SOPS CLI
# Specify SOPS configuration in .sops.yaml
sops -e secret.yaml > secret.enc.yaml
```

### 6. Define KSOPS kustomize Generator

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

### 7. Create the kustomization.yaml

[Read about kustomize plugins](https://kubernetes-sigs.github.io/kustomize/guides/plugins/)

```bash
cat <<EOF > kustomization.yaml
generators:
  - ./secret-generator.yaml
EOF
```

### 8. Build with kustomize 🔑

```bash
# Build with kustomize to verify
# In kustomize v2 and v3 the command is
# kustomize build --enable_alpha_plugins .
kustomize build --enable-alpha-plugins .
```

### Troubleshooting

#### Sanity Checks

- Validate `ksops` is in the `kustomize` plugin path
  - `$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/ksops`

#### Check Existing Issues

Someone might have already encountered your issue.

https://github.com/viaduct-ai/kustomize-sops/issues

## Generator Options

`KSOPS` supports [kustomize exec plugins](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_plugins/#generator-options) annotation based generator options. At the time of writing, the supported annotations are:

- `kustomize.config.k8s.io/needs-hash`
- `kustomize.config.k8s.io/behavior`

For information, read the [kustomize generator options documentation](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/generatorOptions.md).

### Encrypted Secret Overlays w/ Generator Options

Sometimes there is a default secret as part of a project's base manifests, like the [base Argo CD secret](https://github.com/argoproj/argo-cd/blob/master/manifests/base/config/argocd-secret.yaml), which you want to `replace` in your overlay. Other times, you have parts of base secret that are common across different overlays but you want to partially update, or `merge`, changes specific to each overlay as well. You can achieve both of these goals by simply adding the following annotations to your encrypted secrets:

#### Replace a Base Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-secret
  annotations:
    # replace the base secret data/stringData values with these encrypted data/stringData values
    kustomize.config.k8s.io/behavior: replace
type: Opaque
data:
  # Encrypted data here
stringData:
  # Encrypted data here
```

#### Merge/Patch a Base Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-secret
  annotations:
    # merge the base secret data/stringData values with these encrypted data/stringData values
    kustomize.config.k8s.io/behavior: merge
type: Opaque
data:
  # Encrypted data here
stringData:
  # Encrypted data here
```

## Development and Testing

Before developing or testing `KSOPS`, ensure all external [requirements](#requirements) are properly installed.

```bash
# Setup development environment
make setup
```

### Development

`KSOPS` implements the [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugin API in `ksops.go`.

`KSOPS`'s logic is intentionally simple. Given a list of SOPS encrypted Kubernetes manifests, it iterates over each file and decrypts it via SOPS [decrypt](https://godoc.org/go.mozilla.org/sops/decrypt) library. `KSOPS` assumes nothing about the structure of the encrypted resource and relies on [kustomize](https://github.com/kubernetes-sigs/kustomize/) for manifest validation. `KSOPS` expects the encryption key to be accessible. This is important to consider when using `KSOPS` for CI/CD.

### Testing

Testing `KSOPS` requires:

Everything is handled for you by `make test`. Just run it from the repo's root directory:

```bash
make test
```

## Migration from KSOPS v2.x.x to v3.x.x

### Required Changes

- The opt-in `ksops-exec` kind is deprecated. If you are using the `ksops-exec` kind, please migrate to `ksops` once you've upgrade to _v3.x.x_.

### Background

In `KSOPS` _v3.x.x_, the kustomize plugin `ksops` was migrated to an [exec plugin](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_plugins) from a [Go plugin](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#go-plugins) because of simpler installation, dependency management, and package maintenance.

`KSOPS` was originally developed as a [kustomize Go plugin](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#go-plugins). Up until _v2.2.0_ this was the only installation option, but in _v2.2.0_, `KSOPS` introduced an opt-in [exec plugin](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_plugins) under via the `ksops-exec` kind. Now that `KSOPS` is only an [exec plugin](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_plugins), the `ksops-exec` kind is deprecated.

## Argo CD Integration 🤖

`KSOPS` becomes even more powerful when integrated with a CI/CD pipeline. By combining `KSOPS` with [Argo CD](https://github.com/argoproj/argo-cd/), you can manage Kubernetes secrets via the same Git Ops pattern you use to manage the rest of your kubernetes manifests. To integrate `KSOPS` and [Argo CD](https://github.com/argoproj/argo-cd/), you will need to update the Argo CD ConifgMap and create a [strategic merge patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md) or a [custom Argo CD build](https://argoproj.github.io/argo-cd/operator-manual/custom_tools/#byoi-build-your-own-image). As an alternative you can also use the [Argo CD Helm Chart](https://github.com/argoproj/argo-helm/tree/master/charts/argo-cd) with [custom values](#argo-cd-helm-chart-with-custom-tooling). Don't forget to inject any necessary credentials (i.e AWS credentials) when deploying the [Argo CD](https://github.com/argoproj/argo-cd/) + `KSOPS` build!

[KSOPS Docker Image](https://hub.docker.com/r/viaductoss/ksops)

[KSOPS Quay.io Image](https://quay.io/repository/viaductoss/ksops)

### Enable Kustomize Plugins via Argo CD ConfigMap

As of now to allow [Argo CD](https://github.com/argoproj/argo-cd/) to use [kustomize](https://github.com/kubernetes-sigs/kustomize/) plugins you must use the `enable-alpha-plugins` flag. This is configured by the `kustomize.buildOptions` setting in the [Argo CD](https://github.com/argoproj/argo-cd/) ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  labels:
    app.kubernetes.io/name: argocd-cm
    app.kubernetes.io/part-of: argocd
data:
  # For KSOPs versions < v2.5.0, use the old kustomize flag style
  # kustomize.buildOptions: "--enable_alpha_plugins"
  kustomize.buildOptions: "--enable-alpha-plugins"
```

### KSOPS Repo Sever Patch

The simplest way to integrate `KSOPS` with [Argo CD](https://github.com/argoproj/argo-cd/) is with a [strategic merge patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md) on the Argo CD repo server deployment. The patch below uses an init container to build `KSOPS` and volume mount to inject the `KSOPS` plugin and, optionally, override the [kustomize](https://github.com/kubernetes-sigs/kustomize/) executable.

```yaml
# argo-cd-repo-server-ksops-patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      # 1. Define an emptyDir volume which will hold the custom binaries
      volumes:
        - name: custom-tools
          emptyDir: {}
      # 2. Use an init container to download/copy custom binaries into the emptyDir
      initContainers:
        - name: install-ksops
          image: viaductoss/ksops:v3.0.1
          command: ["/bin/sh", "-c"]
          args:
            - echo "Installing KSOPS...";
              mv ksops /custom-tools/;
              mv $GOPATH/bin/kustomize /custom-tools/;
              echo "Done.";
          volumeMounts:
            - mountPath: /custom-tools
              name: custom-tools
      # 3. Volume mount the custom binary to the bin directory (overriding the existing version)
      containers:
        - name: argocd-repo-server
          volumeMounts:
            - mountPath: /usr/local/bin/kustomize
              name: custom-tools
              subPath: kustomize
              # Verify this matches a XDG_CONFIG_HOME=/.config env variable
            - mountPath: /.config/kustomize/plugin/viaduct.ai/v1/ksops/ksops
              name: custom-tools
              subPath: ksops
          # 4. Set the XDG_CONFIG_HOME env variable to allow kustomize to detect the plugin
          env:
            - name: XDG_CONFIG_HOME
              value: /.config
        ## If you use AWS or GCP KMS, don't forget to include the necessary credentials to decrypt the secrets!
        #  - name: AWS_ACCESS_KEY_ID
        #    valueFrom:
        #      secretKeyRef:
        #        name: argocd-aws-credentials
        #        key: accesskey
        #  - name: AWS_SECRET_ACCESS_KEY
        #    valueFrom:
        #      secretKeyRef:
        #        name: argocd-aws-credentials
        #        key: secretkey
```

### Custom Argo CD w/ KSOPS Dockerfile

Alternatively, for more control and faster pod start times you can build a custom docker image.

```Dockerfile
ARG ARGO_CD_VERSION="v1.7.7"
# https://github.com/argoproj/argo-cd/blob/master/Dockerfile
ARG KSOPS_VERSION="v3.0.1"

#--------------------------------------------#
#--------Build KSOPS and Kustomize-----------#
#--------------------------------------------#

FROM viaductoss/ksops:$KSOPS_VERSION as ksops-builder

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

### Argo CD Helm Chart with Custom Tooling

We can setup `KSOPS` custom tooling in the [Argo CD Chart](https://github.com/argoproj/argo-helm/tree/master/charts/argo-cd) with the following values:

```yaml
# Enable Kustomize Alpha Plugins via Argo CD ConfigMap, required for ksops
server:
  config:
    kustomize.buildOptions: "--enable-alpha-plugins"

repoServer:
  # Set the XDG_CONFIG_HOME env variable to allow kustomize to detect the plugin
  env:
    - name: XDG_CONFIG_HOME
      value: /.config

  # Use init containers to configure custom tooling
  # https://argoproj.github.io/argo-cd/operator-manual/custom_tools/
  volumes:
    - name: custom-tools
      emptyDir: {}

  initContainers:
    - name: install-ksops
      image: viaductoss/ksops:v3.0.1
      command: ["/bin/sh", "-c"]
      args:
        - echo "Installing KSOPS...";
          mv ksops /custom-tools/;
          mv $GOPATH/bin/kustomize /custom-tools/;
          echo "Done.";
      volumeMounts:
        - mountPath: /custom-tools
          name: custom-tools
  volumeMounts:
    - mountPath: /usr/local/bin/kustomize
      name: custom-tools
      subPath: kustomize
      # Verify this matches a XDG_CONFIG_HOME=/.config env variable
    - mountPath: /.config/kustomize/plugin/viaduct.ai/v1/ksops/ksops
      name: custom-tools
      subPath: ksops
```
