#!/bin/bash
set -e

PLUGIN_PATH="/usr/local/bin/"

if [[ ! -d "$PLUGIN_PATH" ]]; then
  echo "$PLUGIN_PATH does not exist. Cannot add ksops to PATH."
fi

get_machine_arch () {
    machine_arch=""
    case $(uname -m) in
        i386)    machine_arch="i386" ;;
        i686)    machine_arch="i386" ;;
        x86_64)  machine_arch="x86_64" ;;
        aarch64) machine_arch="arm64" ;;
        arm64)   machine_arch="arm64" ;;
    esac
    echo $machine_arch
}

ARCH=$(get_machine_arch)
OS=""
case $(uname | tr '[:upper:]' '[:lower:]') in
  linux*)
    OS="Linux"
    ;;
  darwin*)
    OS="Darwin"
    ;;
  msys*)
    OS="Windows"
    ;;
  windowsnt*)
    OS="Windows"
    ;;
  *)
    echo "Unknown OS type: $(uname)"
    echo "Please consider contributing to this script to support your OS."
    exit 1
    ;;
esac


echo "Downloading latest release to ksops plugin path"
if [ -x "$(command -v wget)" ]; then
    wget -c https://github.com/viaduct-ai/kustomize-sops/releases/latest/download/ksops_latest_${OS}_${ARCH}.tar.gz -O - | tar -zxvf - ksops
elif [ -x "$(command -v curl)" ]; then
    curl -s -L https://github.com/viaduct-ai/kustomize-sops/releases/latest/download/ksops_latest_${OS}_${ARCH}.tar.gz | tar -zxvf - ksops
else
    echo "This script requires either wget or curl."
    exit 1
fi

if mv ksops $PLUGIN_PATH; then
  echo "Successfully installed ksops"
else
  echo "Failed to move ksops to $PLUGIN_PATH. Maybe you should run this script as root."
  exit 1
fi
