# Should match the version used in https://github.com/LoyaltyNZ/service_base_golang_builder
# See https://golangci-lint.run/usage/install/
LINTER_VERSION = v1.55.0

# Variables needed when building binaries
VERSION := $(shell grep -oE -m 1 '([0-9]+)\.([0-9]+)\.([0-9]+)' CHANGELOG.md )

# To be used for dependencies not installed with gomod
LOCAL_DEPS_INSTALL_LOCATION = /usr/local/bin

.PHONY: clean
clean:
	rm -rf build
	mkdir -p build

.PHONY: deps
deps:
	go env -w "GOPRIVATE=github.com/ildomm/*"
	go mod download

.PHONY: build
build: deps build-loader build-processor

.PHONY: build-loader
build-loader: deps
	# Build the background job binary
	cd cmd/loader && \
		go build -ldflags="-X main.semVer=${VERSION}" \
        -o ../../build/loader

.PHONY: build-processor
build-processor: deps
	# Build the background job binary
	cd cmd/processor && \
		go build -ldflags="-X main.semVer=${VERSION}" \
        -o ../../build/processor

.PHONY: unit-test
unit-test: deps
	go test -tags=testing -count=1 ./...

.PHONY: lint-install
lint-install:
	[ -e ${LOCAL_DEPS_INSTALL_LOCATION}/golangci-lint ] || \
	wget -O- -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b ${LOCAL_DEPS_INSTALL_LOCATION} ${LINTER_VERSION}

.PHONY: lint
lint: deps lint-install
	golangci-lint run

.PHONY: coverage-report
coverage-report: clean deps
	go test -tags=testing ./... \
		-coverprofile=build/cover.out github.com/ildomm/sc_sq_disbursement/...
	go tool cover -html=build/cover.out -o build/coverage.html
	echo "** Coverage is available in build/coverage.html **"

.PHONY: coverage-total
coverage-total: clean deps
	go test -tags=testing ./... \
		-coverprofile=build/cover.out github.com/ildomm/sc_sq_disbursement/...
	go tool cover -func=build/cover.out | grep total
