-include .config/praetor/engine.mk
.PHONY: help test-all coverage-check-all lint-all vet-all vuln-all gosec-all cover-all tidy-all build-all fmt-all golines-all ci-all list-modules clean-all tools-install doc upstream-specs lint-exceptions lint-exceptions-gen

MODULES := $(shell find . -name 'go.mod' -not -path './.workingdir/*' -not -path './.workingdir2/*' -not -path './tools/go/*' -exec dirname {} \; | sort)
GOLANGCI_LINT_VERSION := v2.13.2

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
		cov=$$(cd "$$mod" && go tool cover -func=coverage.out | awk '/^total/{gsub("%","",$$3); print $$3}'); \
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
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	# gofumpt, gci, golines, gosec, govulncheck, apidiff: pinned in tools/go/go.mod (x/tools floor for Go 1.27.2 export data)
	cd tools/go && go install tool

upstream-specs: ## Upstream OpenAPI spec pins: planted cases, then every pinned file against its sha256
	cd tools/go && go vet ./upstreamspec/ && go test -count=1 ./upstreamspec/
	cd tools/go && go run ./upstreamspec check -manifest ../../upstream/specs.yaml

lint-exceptions: ## Dated complexity exceptions in .golangci.yml: planted cases, then the repository config
	cd tools/go && go vet ./lintexceptions/ && go test -count=1 ./lintexceptions/

lint-exceptions-gen: ## Regenerate the dated complexity exceptions from a measured lint run (#141)
	bash tools/lint-exceptions.sh

doc: ## Local godoc server at :6060
	@echo "Starting godoc server at http://localhost:6060"
	@godoc -http=:6060

# Praetor declared verification; existing project recipes remain unchanged.
# Praetor ships one binary under two names; resolve whichever is installed.
PRAETORCTL ?= $(shell command -v praetorctl 2>/dev/null || command -v standardsctl 2>/dev/null || echo praetorctl)
.PHONY: verify-all
verify-all:
	@$(PRAETORCTL) compile-context --verify
	@$(PRAETORCTL) caveman check --configured-sources
	@$(PRAETORCTL) audit
	@$(MAKE) --no-print-directory ci-all

compile-context:
	@$(PRAETORCTL) compile-context

audit:
	@$(PRAETORCTL) audit

# BEGIN praetor documentation gate
.PHONY: docs-lint docs-figures
verify-all: docs-lint docs-figures
docs-lint:
	@node tools/markdownlint/verify.mjs
docs-figures:
	@node tools/figures/build.mjs check
	@node tools/figures/build.mjs sources
# END praetor documentation gate
