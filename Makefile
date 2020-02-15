# Default to installing KSOPS
default: install

.PHONY: install
install:
	./scripts/build-and-install-ksops.sh

.PHONY: kustomize
kustomize:
	./scripts/install-kustomize.sh

.PHONY: test
test: install setup-test-files go-test

.PHONY: setup-test-files
setup-test-files:
	./scripts/setup-test-files.sh

.PHONY: go-test
go-test:
	echo "Running tests..."
	go test -v ./...
