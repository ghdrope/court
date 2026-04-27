# ==== Utility-style Makefile for Court Project ====

# Use bash for all systems consistency
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# ==== Variables ====
ARTIFACTS_DIR := $(PWD)/.reports/pipelines
BIN_DIR := .bin
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
CACHE_DIR := $(PWD)/.cache
CI_DOCKER_IMAGES := \
	alpine:3.23 \
	golang:1.26.1 \
	golang:1.26.1-alpine
CI_DOCKER_IMAGES_ALLOWLIST := \
	alpine:3.23 \
	golang:1.26.1 \
	golang:1.26.1-alpine
CI_VULN_REPORT_DIR := $(ARTIFACTS_DIR)/security/images
CI_VULN_SCANNER := trivy
CI_VULN_SEVERITIES := HIGH,CRITICAL
COMPONENT ?=
GIT_COMMIT ?= $(shell git rev-parse HEAD)
GOCACHE_DIR := go-build
GOVULNCHECK_ARTIFACT := govulncheck-report.json
PREFIX ?= /usr/local
PROJECT_NAME := court
SDLC_ARTIFACTS_DIR := SDLC
SECURITY_ARTIFACTS_DIR := security
UNIT_TEST_OUT_ARTIFACT := coverage.out
UNIT_TEST_XML_ARTIFACT := coverage.xml
VERSION ?= development

# Export coverage variables to shell
export MIN_COVERAGE=45.0


# ==== Convenience targets ====
##@Convenience

.PHONY: build-officer build-court
build-officer: # Build officer component
	$(MAKE) build COMPONENT=officer

build-court: ## Build court component
	$(MAKE) build COMPONENT=court


# ==== Guards ====
.PHONY: check-component
check-component: # Ensure COMPONENT variable is set
ifndef COMPONENT
	$(error COMPONENT is not set. Usage: make build-component COMPONENT=<name>)
endif


# ==== Clean ====
.PHONY: clean
clean: ## Complete clean (using all clean available targets)
	@echo "[TASK] Complete clean"
	$(MAKE) clean-build
	@echo "✅ Clean completed successfully"

.PHONY: clean-build
clean-build: ## Clean build artifacts, caches, and reports
	@echo "[TASK] Clean build artifacts and caches"
	@rm -rf .bin .cache .reports
	@go clean -testcache
	@echo "✅ Build clean completed successfully"


# ==== Security ====
.PHONY: check-vulnerability
check-vulnerability: ## Run Go vulnerability check
	@echo "[TASK] Running vulnerability check"
	@mkdir -p "$(ARTIFACTS_DIR)/$(SECURITY_ARTIFACTS_DIR)"
	@REPORT="$(ARTIFACTS_DIR)/$(SECURITY_ARTIFACTS_DIR)/$(GOVULNCHECK_ARTIFACT)"; \
	go run golang.org/x/vuln/cmd/govulncheck@latest ./... > $$REPORT; \
	\
	if grep -q '"finding"' $$REPORT; then \
		echo "❌ Vulnerabilities found:"; \
		jq -r 'select(.finding != null) | .finding.osv' $$REPORT; \
		exit 1; \
	else \
		echo "✅ No vulnerabilities found"; \
	fi

.PHONY: check-cicd-vulnerability
check-cicd-vulnerability: ## Run CI images vulnerability check
	@echo "[TASK] Scanning CI Docker images for vulnerabilities"
	@command -v trivy >/dev/null 2>&1 || { echo "❌ trivy is not installed"; exit 1; }

	@mkdir -p "$(CI_VULN_REPORT_DIR)"

	@FAILED_IMAGES=""; \
	for IMAGE in $(CI_DOCKER_IMAGES); do \
		echo "🔍 Scanning $$IMAGE"; \
		REPORT_FILE="$(CI_VULN_REPORT_DIR)/$$(echo $$IMAGE | tr '/:' '__')__vuln.json"; \
		mkdir -p "$$(dirname "$$REPORT_FILE")"; \
		trivy image \
			--severity $(CI_VULN_SEVERITIES) \
			--no-progress \
			--timeout 10m \
			--exit-code 1 \
			--format json \
			--output "$$REPORT_FILE" \
			"$$IMAGE" && echo "✅ $$IMAGE is clean" || { \
				if echo "$(CI_DOCKER_IMAGES_ALLOWLIST)" | grep -qw "$$IMAGE"; then \
					echo "⚠️ $$IMAGE has vulnerabilities but is in the allowlist — ignoring"; \
				else \
					FAILED_IMAGES="$$FAILED_IMAGES $$IMAGE"; \
				fi \
			}; \
		echo "📄 Report saved to: $$REPORT_FILE"; \
	done; \
	if [ -n "$$FAILED_IMAGES" ]; then \
		echo ""; \
		echo "🚨 CI Docker image vulnerability scan FAILED"; \
		echo "Images with HIGH or CRITICAL vulnerabilities:"; \
		for IMG in $$FAILED_IMAGES; do echo "  - $$IMG"; done; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ All CI Docker images passed vulnerability scanning"; \
	fi


# === Quality ===
.PHONY: format-check
format-check: ## Check code formatting
	@echo "[TASK] Checking code formatting"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Formatting issues found:"; \
		gofmt -l ./court; \
		exit 1; \
	else \
		echo "✅ Code formatting is correct"; \
	fi

.PHONY: lint
lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { echo "❌ golangci-lint is not installed"; exit 1; }
	@echo "[TASK] Running golangci-lint for component"
	@if golangci-lint run ./...; then \
		echo "✅ Lint passed successfully"; \
	else \
		echo "❌ Lint failed"; \
		exit 1; \
	fi


# ==== Build lifecycle ====
.PHONY: build
build: check-component # Build a single component binary
	@echo "[TASK] Building component $(COMPONENT)"

	@mkdir -p "$(CACHE_DIR)/$(GOCACHE_DIR)" "$(BIN_DIR)/"
	@export GOCACHE="$(CACHE_DIR)/$(GOCACHE_DIR)" && \
	\
	go mod download && \
	echo "🔨 Building binary" && \
	go build -ldflags "\
		-X 'court/pkg/version.Version=$(VERSION)' \
		-X 'court/pkg/version.GitCommit=$(GIT_COMMIT)' \
		-X 'court/pkg/version.BuildDate=$(BUILD_DATE)'" \
		-o "$(BIN_DIR)/$(PROJECT_NAME)-$(COMPONENT)" "$(PWD)/cmd/$(COMPONENT)"; \
	echo "✅ Build completed successfully"; \


# ==== Tests ====
.PHONY: test-unit
test-unit: build ## Run unit tests with coverage enforcement
	@echo "[TASK] Running unit tests"
	@mkdir -p "$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)"

	PACKAGES=$$(go list ./internal/... ./pkg/... | grep -v '/tests' | grep -v 'testhelper' | grep -v '^court/$$') && \
	echo "$$PACKAGES" && \
	\
	COVERAGE_MIN=$$MIN_COVERAGE; \
	\
	go test -v -coverprofile="$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)/court-$(UNIT_TEST_OUT_ARTIFACT)" $$PACKAGES && \
	\
	COVERAGE_ACTUAL=$$(go tool cover -func="$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)/court-$(UNIT_TEST_OUT_ARTIFACT)" | grep total: | awk '{print substr($$3,1,length($$3)-1)}') && \
	if awk "BEGIN {exit !($$COVERAGE_ACTUAL >= $$COVERAGE_MIN)}"; then \
		echo "📊 Total Coverage $$COVERAGE_ACTUAL% >= Minimum Coverage $$COVERAGE_MIN%"; \
	else \
		echo "❌ Total Coverage $$COVERAGE_ACTUAL% < Minimum Coverage $$COVERAGE_MIN%"; \
		exit 1; \
	fi && \
	\
	if command -v gocover-cobertura >/dev/null 2>&1; then \
		gocover-cobertura < "$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)/court-$(UNIT_TEST_OUT_ARTIFACT)" > "$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)/court-$(UNIT_TEST_XML_ARTIFACT)" && \
		echo "📝 Cobertura report generated: '$(ARTIFACTS_DIR)/$(SDLC_ARTIFACTS_DIR)/court-$(UNIT_TEST_XML_ARTIFACT)'"; \
	else \
		echo "⚠️ gocover-cobertura not found, skipping Cobertura report"; \
	fi 

	@echo "✅ Unit tests completed successfully"


# ==== Help ====
.PHONY: help
help: ## help: list available targets
	@echo "Available targets:"
	@gawk '\
		/^##@Convenience/ {in_convenience=1; next} \
		/^##@/ {in_convenience=0} \
		in_convenience && /^[a-zA-Z0-9_-]+:.*?##/ { \
			match($$0, /^([a-zA-Z0-9_-]+):/, arr); \
			target = arr[1]; \
			sub(/^.*## /,"",$$0); \
			printf "  \033[36m%-30s\033[0m %s\n", target, $$0 \
		} \
	' $(MAKEFILE_LIST)
