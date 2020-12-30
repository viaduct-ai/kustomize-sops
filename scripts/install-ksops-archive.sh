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

ARCH=$(uname -m)
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
wget -c https://github.com/viaduct-ai/kustomize-sops/releases/latest/download/ksops_latest_${OS}_${ARCH}.tar.gz -O - | tar -xz -C $PLUGIN_PATH
echo "Successfully installed ksops"
