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

PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/$PLUGIN_NAME/"

mkdir -p $PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
cp $PLUGIN_NAME $PLUGIN_PATH

# ------------------------
# Deprecated ksops-exec Plugin
# Please migrate to ksops if you are using ksops-exec
# ------------------------
DEPRECATED_EXEC_PLUGIN_NAME="ksops-exec"

DEPRECATED_PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/$DEPRECATED_EXEC_PLUGIN_NAME"

mkdir -p $DEPRECATED_PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $DEPRECATED_PLUGIN_PATH/$DEPRECATED_EXEC_PLUGIN_NAME"
cp $PLUGIN_NAME $DEPRECATED_PLUGIN_PATH/$DEPRECATED_EXEC_PLUGIN_NAME
