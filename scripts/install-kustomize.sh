#!/bin/bash
set -e

KUSTOMIZE="kustomize"

function install_kustomize() {
  echo "Installing $KUSTOMIZE..."
  KUSTOMIZE_MAJOR_VERSION='v5'
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

  # Determine the correct installation path
  GO_PATH=$(go env GOPATH)
  KUSTOMIZE_PATH="$GO_PATH/bin/$KUSTOMIZE"

  # Ensure the bin directory exists
  mkdir -p "$GO_PATH/bin"

  # Build kustomize in a temporary directory to avoid module conflicts
  TEMP_DIR=$(mktemp -d)
  cd "$TEMP_DIR"

  # Initialize a temporary module and set Go environment
  go mod init temp-kustomize-build
  export GOCACHE="$TEMP_DIR/.cache"
  export GOMODCACHE="$TEMP_DIR/pkg/mod"
  GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/$KUSTOMIZE_MAJOR_VERSION@$KUSTOMIZE_VERSION
  GO111MODULE=on go build -ldflags "${LDFLAGS}" -o "$KUSTOMIZE_PATH" sigs.k8s.io/kustomize/kustomize/$KUSTOMIZE_MAJOR_VERSION

  echo "kustomize installed at $KUSTOMIZE_PATH"

  $KUSTOMIZE_PATH version
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
    rm "$KUSTOMIZE_EXEC"

    # Install
    install_kustomize
  fi
else
    # Install
    install_kustomize
fi
