GOACC_PKG := github.com/ory/go-acc
GOACC_COMMIT := 29bd61f765d44842832fb2cdc01d9eab58d1cd3e

GO_BIN := ${GOPATH}/bin
GOACC_BIN := $(GO_BIN)/go-acc

DEPINSTALL := cd /tmp && go install -v

GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all install-deps vet unit-test test fmt lint test-cov install-depu module-updates install help

all: install-deps vet test

$(GOACC_BIN):
	$(DEPINSTALL) $(GOACC_PKG)@$(GOACC_COMMIT)

install-deps:
	go mod download
	go mod verify

vet:
	go vet github.com/piprate/metalocker/...

unit-test: vet
	go test github.com/piprate/metalocker/...

test: unit-test

fmt:
	gofmt -s -w ${GOFILES_NOVENDOR}

lint:
	golangci-lint run

test-cov: $(GOACC_BIN)
	$(GOACC_BIN) ./...

install-depu:
	go install github.com/kevwan/depu@latest

module-updates:
	depu

install:
	GOBIN=`pwd`/bin/ go install -v github.com/piprate/metalocker/cmd/...

help:
	@echo ''
	@echo ' Targets:'
	@echo '--------------------------------------------------'
	@echo ' all              - Run everything                '
	@echo ' fmt              - Format code                   '
	@echo ' vet              - Run vet                       '
	@echo ' test             - Run all tests                 '
	@echo ' unit-test        - Run unit tests                '
	@echo ' lint             - Run golangci-lint             '
	@echo ' test-cov         - Run all tests + coverage      '
	@echo '--------------------------------------------------'
	@echo ''
