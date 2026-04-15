.PHONY: help test-all lint-all vet-all vuln-all gosec-all cover-all tidy-all build-all fmt-all golines-all ci-all list-modules clean-all tools-install doc

MODULES := $(shell find . -name 'go.mod' -not -path './.workingdir/*' -not -path './.workingdir2/*' -exec dirname {} \;)

help:  ## Show this help
	@awk 'BEGIN{FS=":.*?## "} /^[a-zA-Z_-]+:.*?## /{printf "  \033[36m%-16s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

list-modules: ## Print every discovered go.mod directory
	@printf '%s\n' $(MODULES)

test-all: ## go test -race + coverage, all modules
	@for mod in $(MODULES); do \
		echo "==> Testing $$mod"; \
		(cd $$mod && go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...) || exit 1; \
	done

cover-all: test-all ## Generate coverage.html per module
	@for mod in $(MODULES); do \
		(cd $$mod && go tool cover -html=coverage.out -o coverage.html) || true; \
	done

lint-all: ## golangci-lint run, all modules
	@for mod in $(MODULES); do \
		echo "==> Linting $$mod"; \
		(cd $$mod && golangci-lint run --config=$(CURDIR)/.golangci.yml ./...) || exit 1; \
	done

vet-all: ## go vet, all modules
	@for mod in $(MODULES); do \
		echo "==> Vetting $$mod"; \
		(cd $$mod && go vet ./...) || exit 1; \
	done

vuln-all: ## govulncheck, all modules
	@for mod in $(MODULES); do \
		echo "==> govulncheck $$mod"; \
		(cd $$mod && govulncheck ./...) || exit 1; \
	done

gosec-all: ## gosec, all modules
	@for mod in $(MODULES); do \
		echo "==> gosec $$mod"; \
		(cd $$mod && gosec -quiet -exclude-generated ./...) || exit 1; \
	done

tidy-all: ## go mod tidy, all modules
	@for mod in $(MODULES); do \
		echo "==> Tidying $$mod"; \
		(cd $$mod && go mod tidy) || exit 1; \
	done

fmt-all: ## gofumpt + gci, all modules
	@for mod in $(MODULES); do \
		echo "==> Formatting $$mod"; \
		(cd $$mod && gofumpt -w . && gci write --skip-generated -s standard -s default -s 'prefix(github.com/golusoris/goenvoy)' .) || exit 1; \
	done

golines-all: ## golines -m 120, all modules (opt-in, not a CI gate)
	@for mod in $(MODULES); do \
		echo "==> golines $$mod"; \
		(cd $$mod && golines -m 120 -w .) || exit 1; \
	done

build-all: ## go build, all modules
	@for mod in $(MODULES); do \
		echo "==> Building $$mod"; \
		(cd $$mod && go build ./...) || exit 1; \
	done

clean-all: ## Remove coverage artefacts
	@for mod in $(MODULES); do \
		rm -f $$mod/coverage.out $$mod/coverage.html; \
	done

ci-all: lint-all vet-all gosec-all vuln-all test-all ## Full local CI — matches GitHub Actions gates

tools-install: ## Install gofumpt, gci, golines, gosec, govulncheck, apidiff
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/segmentio/golines@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/exp/cmd/apidiff@latest

doc: ## Local godoc server at :6060
	@echo "Starting godoc server at http://localhost:6060"
	@godoc -http=:6060
