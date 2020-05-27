#!/bin/bash
set -e

KUSTOMIZE="kustomize"

function install_kustomize() {
  echo "Installing $KUSTOMIZE..."
  GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@v3.6.1

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

