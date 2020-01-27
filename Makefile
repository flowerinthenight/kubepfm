VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || cat $(CURDIR)/.version 2> /dev/null || echo v0)
BLDVER = module:$(MODULE),version:$(VERSION),build:$(shell date +"%Y%m%d.%H%M%S.%N.%z")
BASE = $(CURDIR)
MODULE = kubepfm

.PHONY: all $(MODULE)
all: $(MODULE)

$(MODULE):| $(BASE)
	@GO111MODULE=on GOFLAGS=-mod=vendor go install -v

$(BASE):
	@mkdir -p $(dir $@)
