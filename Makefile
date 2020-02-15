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

# https://vincent.bernat.ch/en/blog/2019-makefile-build-golang
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


################################################################################
# Git Hooks
################################################################################
## Git hooks to validate worktree is clean before commit/push
.git/hooks/pre-push: Makefile
	# Create Git pre-push hook
	echo 'make pre-push' > .git/hooks/pre-push
	chmod +x .git/hooks/pre-push

## Git hooks to validate worktree is clean before commit/push
.git/hooks/pre-commit: Makefile
	# Create Git pre-commit hook
	echo 'make pre-commit' > .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

.PHONY: must-be-clean
must-be-clean:
	# Check everything has been committed to Git
	@if [ "$(GIT_TREE_STATE)" != "clean" ]; then echo 'git tree state is $(GIT_TREE_STATE)' ; exit 1; fi

.PHONY: pre-push
pre-push: must-be-clean pre-commit must-be-clean

.PHONY: pre-commit
pre-commit: download-dependencies lint test

