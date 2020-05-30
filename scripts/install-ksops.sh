#!/bin/bash
set -e

if [[ -z "$XDG_CONFIG_HOME" ]]; then
  echo "HM $HOME"
  echo "You must define XDG_CONFIG_HOME to use a kustomize plugin"
  echo "Add 'export XDG_CONFIG_HOME=\$HOME/.config' to your .bashrc"
  exit 1
fi


PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/"
# Unclear why the kustomize test harness looks for the plugin relative to the current path
# https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/plugins/utils/utils.go#L22
TEST_PLUGIN_PATH="$HOME/sigs.k8s.io/kustomize/plugin/viaduct.ai/v1/ksops/"

PLUGIN_NAME="ksops.so"

mkdir -p $PLUGIN_PATH
mkdir -p $TEST_PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying executable plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
cp $PLUGIN_NAME $PLUGIN_PATH

echo "Copying executable plugin to the test kustomize plugin path..."
echo "cp $PLUGIN_NAME $TEST_PLUGIN_PATH"
cp $PLUGIN_NAME $TEST_PLUGIN_PATH
