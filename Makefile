PLUGIN_NAME="ksops.so"

# Default to installing KSOPS
default: install

.PHONY: install
install: build install-plugin

.PHONY: install-plugin
install-plugin:
	./scripts/install-ksops.sh

.PHONY: build
build: build-plugin

.PHONY: build-plugin
build-plugin:
	go build -buildmode plugin -o $(PLUGIN_NAME)

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

.PHONY: download-dependencies
download-dependencies:
	go mod download
	go mod tidy

BIN = $(CURDIR)/bin
$(BIN):
		@mkdir -p $@
$(BIN)/%: | $(BIN)
		@tmp=$$(mktemp -d); \
       env GO111MODULE=off GOPATH=$$tmp GOBIN=$(BIN) go get $(PACKAGE) \
        || ret=$$?; \
       rm -rf $$tmp ; exit $$ret

$(BIN)/golint: PACKAGE=golang.org/x/lint/golint

GOLINT = $(BIN)/golint
lint: | $(GOLINT)
		$(GOLINT) -set_exit_status ./...
