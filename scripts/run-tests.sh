set -e

sh ./scripts/build-and-install-ksops.sh

# Setup tests files
sh ./scripts/setup-test-files.sh 

# Run the tests
echo "Running tests..."
go test -v
