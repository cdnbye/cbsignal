# Makefile用于交叉编译
SHELL := /bin/bash
BASEDIR = $(shell pwd)

# Go parameters
GOCMD=go
GOENV=GOOS=linux GOARCH=amd64
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME_SIGNALER=cbsignal

main:
	@echo "------build signaler--------"
	$(GOENV) $(GOBUILD) -o $(BINARY_NAME_SIGNALER) -v main.go
test:
	$(GOTEST) -v ./...
clean:
	@echo "------clean--------"
	rm -f $(BINARY_NAME_SIGNALER)
help:
	@echo "make - compile the source code"
	@echo "make clean - remove binary file and vim swp files"
.PHONY: clean help



