PACKAGE     =  github.com/saltosystems/winrt-go
PKG         ?= ./...
APP         ?= winrt-go-gen
BUILD_TAGS  ?= 

include .go-builder/Makefile

.PHONY: prepare
prepare: $(prepare_targets)

.PHONY: sanity-check
sanity-check: $(sanity_check_targets) check-generated

.PHONY: build
build: $(build_targets)

.PHONY: test
test: $(test_targets) go-test

.PHONY: release
release: $(release_targets)

.PHONY: clean
clean: $(clean_targets)

.PHONY: gen-files
gen-files:
	rm -rf $(CURDIR)/windows
	go generate github.com/saltosystems/winrt-go/...

.PHONY: check-generated
check-generated: export WINRT_GO_GEN_VALIDATE=1
check-generated:
	 go generate github.com/saltosystems/winrt-go/...

.PHONY: go-test
go-test:
	go test github.com/saltosystems/winrt-go/...
