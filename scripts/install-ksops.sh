#!/bin/bash
set -e

# Require $XDG_CONFIG_HOME to be set
if [[ -z "$XDG_CONFIG_HOME" ]]; then
  echo "You must define XDG_CONFIG_HOME to use a kustomize plugin"
  echo "Add 'export XDG_CONFIG_HOME=\$HOME/.config' to your .bashrc or .zshrc"
  exit 1
fi

PLUGIN_NAME="ksops"

# ------------------------
# ksops Plugin
# ------------------------

PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/"

mkdir -p $PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
cp $PLUGIN_NAME $PLUGIN_PATH
