#           __                                      
#     _____/ /_  __  ______ _____  __  ____  ______ 
#    / ___/ __ \/ / / / __ `/ __ \/ / / / / / / __ \
#   / /__/ / / / /_/ / /_/ / / / / /_/ / /_/ / / / /
#   \___/_/ /_/\__,_/\__,_/_/ /_/\__, /\__,_/_/ /_/ 
#                               /____/              
#   ================================================
#   chuanyun.io quasimodo program.


GO           ?= GO15VENDOREXPERIMENT=1 go
GOPATH 		 := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))

STATICCHECK  ?= $(GOPATH)/bin/staticcheck
pkgs          = $(shell $(GO) list ./... | grep -v /vendor/)

BINARY  	  = quasimodo
DATE         ?= $(shell date +%FT%T%z)
VERSION      ?= $(shell git describe --tags --always --dirty="-dev" --match=v* 2> /dev/null || echo v0)

LDFLAGS       = -ldflags "-X main.GitTag=${VERSION} -X main.BuildTime=${DATE}"


all: format vet build
	@echo
	@echo "Build complete."
	@echo "Don't forget to run 'make test'."
	@echo

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)
	
vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

staticcheck: $(STATICCHECK)
	@echo ">> running staticcheck"
	@$(STATICCHECK) $(pkgs)

build: 
	@echo ">> building binaries"
	@$(GO) build ${LDFLAGS} -o $(CURDIR)/target/bin/${BINARY}

test:
	@echo ">> running tests"
	@$(GO) test -short -race $(pkgs)


$(GOPATH)/bin/staticcheck:
	@GOOS= GOARCH= $(GO) get -u honnef.co/go/tools/cmd/staticcheck


.PHONY: all format vet build test

.PHONY: $(GOPATH)/bin/staticcheck

