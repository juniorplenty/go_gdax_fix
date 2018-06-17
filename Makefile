SHELL := /bin/bash

all: run test

.PHONY: run
run:
	rm -fr logs && . .env && go run cmd/go_gdax_fix/go_gdax_fix.go

.PHONY: run_race
run_race:
	rm -fr logs && . .env && go run -race cmd/go_gdax_fix/go_gdax_fix.go

.PHONY: test
test:
	rm -fr logs && . .env && go test -v -race ./...
