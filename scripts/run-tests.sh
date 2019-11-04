set -e

PLUGIN_PATH="~/.config/kustomize/plugin/viaduct.ai/v1/ksops/"
PLUGIN_NAME="ksops.so"

# Build Go plugin
echo "Building KSOPS Go plugin..."
go build -buildmode plugin -o $PLUGIN_NAME

# Make the plugin available to kustomize 
echo "Copying executable plugin to the kustomize plugin path..."
echo "cp $PLUGIN_NAME $PLUGIN_PATH"
mkdir -p $PLUGIN_PATH
cp $PLUGIN_NAME $PLUGIN_PATH

# Setup tests files
echo "Generating test files..."
sh ./scripts/setup-test-files.sh 

# Run the tests
echo "Running tests..."
go test
