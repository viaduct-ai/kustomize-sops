set -e

XDG_CONFIG_HOME="./test/.config"
PLUGIN_PATH="$XDG_CONFIG_HOME/kustomize/plugin/viaduct.ai/v1/ksops/"
PLUGIN_NAME="ksops.so"

mkdir -p $PLUGIN_PATH

# Build Go plugin
echo "Building KSOPS Go plugin..."
go build -buildmode plugin -o $PLUGIN_NAME

# Make the plugin available to kustomize 
echo "Copying executable plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
cp $PLUGIN_NAME $PLUGIN_PATH

# Setup tests files
echo "Generating test files..."
sh ./scripts/setup-test-files.sh 

# Run the tests
echo "Running tests..."
go test
