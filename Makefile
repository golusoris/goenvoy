.PHONY: help test-all coverage-check-all lint-all vet-all vuln-all gosec-all cover-all tidy-all build-all fmt-all golines-all ci-all list-modules clean-all tools-install doc

MODULES := $(shell find . -name 'go.mod' -not -path './.workingdir/*' -not -path './.workingdir2/*' -exec dirname {} \; | sort)
GOLANGCI_LINT_VERSION := v2.12.2
GOSEC_VERSION := v2.27.1
GOVULNCHECK_VERSION := v1.4.0
APIDIFF_VERSION := v0.0.0-20260410095643-746e56fc9e2f

help:  ## Show this help
	@awk 'BEGIN{FS=":.*?## "} /^[a-zA-Z_-]+:.*?## /{printf "  \033[36m%-16s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

list-modules: ## Print every discovered go.mod directory
	@printf '%s\n' $(MODULES)

test-all: ## go test -race + coverage, all modules
	@for mod in $(MODULES); do \
		echo "==> Testing $$mod"; \
		(cd $$mod && go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...) || exit 1; \
	done

coverage-check-all: ## Enforce per-module coverage thresholds from tools/coverage-thresholds.json
	@command -v jq >/dev/null 2>&1 || { echo "jq is required for coverage-check-all"; exit 1; }
	@for mod in $(MODULES); do \
		key=$${mod#./}; \
		file="$$mod/coverage.out"; \
		if [ ! -s "$$file" ]; then \
			echo "==> Coverage $$key: no coverage.out, skipping"; \
			continue; \
		fi; \
		threshold=$$(jq -r --arg k "$$key" '.overrides[$$k] // .default' tools/coverage-thresholds.json); \
		cov=$$(go tool cover -func="$$file" | awk '/^total/{gsub("%","",$$3); print $$3}'); \
		if [ -z "$$cov" ]; then \
			echo "==> Coverage $$key: no executable statements, skipping"; \
			continue; \
		fi; \
		if [ "$$cov" = "0.0" ]; then \
			cov_lines=$$(grep -cv '^mode:' "$$file" || true); \
			if [ "$$cov_lines" -eq 0 ]; then \
				echo "==> Coverage $$key: no executable statements, skipping"; \
				continue; \
			fi; \
		fi; \
		echo "==> Coverage $$key: $$cov% (threshold $$threshold%)"; \
		if awk -v c="$$cov" -v t="$$threshold" 'BEGIN { exit !(c+0 < t+0) }'; then \
			echo "coverage failure: $$key $$cov% < $$threshold%"; \
			exit 1; \
		fi; \
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

ci-all: lint-all vet-all gosec-all vuln-all test-all coverage-check-all build-all ## Full local CI - matches GitHub Actions gates

tools-install: ## Install pinned local tools used by CI and release checks
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/segmentio/golines@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install golang.org/x/exp/cmd/apidiff@$(APIDIFF_VERSION)

doc: ## Local godoc server at :6060
	@echo "Starting godoc server at http://localhost:6060"
	@godoc -http=:6060
