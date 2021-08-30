#!/bin/bash
set -e

KUSTOMIZE="kustomize"

function install_kustomize() {
  echo "Installing $KUSTOMIZE..."
  KUSTOMIZE_MAJOR_VERSION='v4'
  KUSTOMIZE_VERSION="${KUSTOMIZE_MAJOR_VERSION}.3.0"
  BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  KSOPS_VERSION=$(git rev-parse HEAD)
  KSOPS_TAG=$(git describe --exact-match --tags HEAD 2>/dev/null || true )
  LDFLAGS="-X sigs.k8s.io/kustomize/api/provenance.buildDate=${BUILD_DATE}"
  LDFLAGS+=" -X sigs.k8s.io/kustomize/api/provenance.gitCommit=${KUSTOMIZE_MAJOR_VERSION}@${KUSTOMIZE_VERSION}"
  if [ ! -z $KSOPS_TAG ]; then
    KSOPS_VERSION=$KSOPS_TAG
  fi
  LDFLAGS+=" -X sigs.k8s.io/kustomize/api/provenance.version=${KUSTOMIZE_VERSION}+ksops.${KSOPS_VERSION}"
  GO111MODULE=on go get -ldflags "${LDFLAGS}" sigs.k8s.io/kustomize/kustomize/$KUSTOMIZE_MAJOR_VERSION@$KUSTOMIZE_VERSION

  echo "Successfully installed $KUSTOMIZE!"
  kustomize version
}

if [ -x "$(command -v $KUSTOMIZE)" ]; then
  KUSTOMIZE_EXEC=$(command -v $KUSTOMIZE)

  echo "WARNING: Found an existing installation of $KUSTOMIZE at $KUSTOMIZE_EXEC"
  read -p "Please confirm you want to reinstall $KUSTOMIZE (y/n): " -n 1 -r
  echo 
  if [[ $REPLY =~ ^[Yy]$ ]]
  then
    # Remove existing kustomize executable
    echo "Removing existing $KUSTOMIZE executable..."
    echo "rm $KUSTOMIZE_EXEC"
    rm $KUSTOMIZE_EXEC

    # Install
    install_kustomize
  fi
else
    # Install
    install_kustomize
fi
