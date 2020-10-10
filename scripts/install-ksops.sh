#!/bin/bash
set -e

# Require $XDG_CONFIG_HOME to be set
if [[ -z "$XDG_CONFIG_HOME" ]]; then
  echo "You must define XDG_CONFIG_HOME to use a kustomize plugin"
  echo "Add 'export XDG_CONFIG_HOME=\$HOME/.config' to your .bashrc"
  exit 1
fi

# mode parameter
mode=${mode}

# extract parameters
while [ $# -gt 0 ]; do

   if [[ $1 == *"--"* ]]; then
        param="${1/--/}"
        declare $param="$2"
        # echo $1 $2 // Optional to see the parameter:value result
   fi

  shift
done

PLUGIN_NAME="ksops.so"

# if mode is exec install exec plugin under both /ksops/ /ksops-exec/
if [ $mode == "exec" ]; then
  PLUGIN_NAME="ksops"
fi


# ------------------------
# ksops Plugin
# ------------------------

PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/"
# Unclear why the kustomize test harness looks for the plugin relative to the current path
# https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/plugins/utils/utils.go#L22
TEST_PLUGIN_PATH="$HOME/sigs.k8s.io/kustomize/plugin/viaduct.ai/v1/ksops/"


mkdir -p $PLUGIN_PATH
mkdir -p $TEST_PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
cp $PLUGIN_NAME $PLUGIN_PATH

echo "Copying plugin to the test kustomize plugin path..."
echo "cp $PLUGIN_NAME $TEST_PLUGIN_PATH"
cp $PLUGIN_NAME $TEST_PLUGIN_PATH

# ------------------------
# ksops-exec Plugin
# ------------------------

EXEC_PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops-exec"
EXEC_TEST_PLUGIN_PATH="$HOME/sigs.k8s.io/kustomize/plugin/viaduct.ai/v1/ksops-exec"

EXEC_PLUGIN_NAME="ksops"
EXEC_PLUGIN_KIND="ksops-exec"

mkdir -p $EXEC_PLUGIN_PATH
mkdir -p $EXEC_TEST_PLUGIN_PATH

# Make the plugin available to kustomize 
echo "Copying exec plugin to the kustomize plugin path..."
echo "cp $EXEC_PLUGIN_NAME $EXEC_PLUGIN_PATH/$EXEC_PLUGIN_KIND"
cp $EXEC_PLUGIN_NAME "$EXEC_PLUGIN_PATH/$EXEC_PLUGIN_KIND"

echo "Copying exec plugin to the test kustomize plugin path..."
echo "cp $EXEC_PLUGIN_NAME $EXEC_TEST_PLUGIN_PATH/$EXEC_PLUGIN_KIND"
cp $EXEC_PLUGIN_NAME "$EXEC_TEST_PLUGIN_PATH/$EXEC_PLUGIN_KIND"
