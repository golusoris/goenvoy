.PHONY: test-all lint-all tidy-all vet-all build-all doc fmt-all

# Find all directories containing a go.mod file
MODULES := $(shell find . -name 'go.mod' -not -path './.workingdir/*' -exec dirname {} \;)

# Run tests across all modules with race detector
test-all:
	@for mod in $(MODULES); do \
		echo "==> Testing $$mod"; \
		(cd $$mod && go test -race -coverprofile=coverage.out ./...) || exit 1; \
	done

# Run golangci-lint across all modules
lint-all:
	@for mod in $(MODULES); do \
		echo "==> Linting $$mod"; \
		(cd $$mod && golangci-lint run ./...) || exit 1; \
	done

# Run go vet across all modules
vet-all:
	@for mod in $(MODULES); do \
		echo "==> Vetting $$mod"; \
		(cd $$mod && go vet ./...) || exit 1; \
	done

# Run go mod tidy across all modules
tidy-all:
	@for mod in $(MODULES); do \
		echo "==> Tidying $$mod"; \
		(cd $$mod && go mod tidy) || exit 1; \
	done

# Build all modules
build-all:
	@for mod in $(MODULES); do \
		echo "==> Building $$mod"; \
		(cd $$mod && go build ./...) || exit 1; \
	done

# Format all modules
fmt-all:
	@for mod in $(MODULES); do \
		echo "==> Formatting $$mod"; \
		(cd $$mod && gofmt -s -w .) || exit 1; \
	done

# Start local godoc server
doc:
	@echo "Starting godoc server at http://localhost:6060"
	@godoc -http=:6060

# List all modules
list-modules:
	@for mod in $(MODULES); do \
		echo "$$mod"; \
	done
