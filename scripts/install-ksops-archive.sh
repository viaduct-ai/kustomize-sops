#!/bin/bash
set -e

# Require $XDG_CONFIG_HOME to be set
if [[ -z "$XDG_CONFIG_HOME" ]]; then
  echo "You must define XDG_CONFIG_HOME to use a kustomize plugin"
  echo "Add 'export XDG_CONFIG_HOME=\$HOME/.config' to your .bashrc or .zshrc"
  exit 1
fi


PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/"


echo "Verify ksops plugin directory exists and is empty"
rm -rf $PLUGIN_PATH || true
mkdir -p $PLUGIN_PATH

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
    wget -c https://github.com/viaduct-ai/kustomize-sops/releases/latest/download/ksops_latest_${OS}_${ARCH}.tar.gz -O - | tar -xz -C $PLUGIN_PATH
elif [ -x "$(command -v curl)" ]; then
    curl -s -L https://github.com/viaduct-ai/kustomize-sops/releases/latest/download/ksops_latest_${OS}_${ARCH}.tar.gz | tar -xz -C $PLUGIN_PATH
else
    echo "This script requires either wget or curl."
    exit 1
fi
echo "Successfully installed ksops"
