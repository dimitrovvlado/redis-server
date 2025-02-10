# Set an output prefix, which is the local directory if not specified
PREFIX ?= $(shell pwd)
REDIS := /usr/local/opt/redis/bin/redis-server
GO := go

.PHONY: run-redis
run-redis:
	@echo "+ $@"
	$(REDIS) ./redis.conf

.PHONY: run-server
run-server:
	@echo "+ $@"
	$(GO) run cmd/server/main.go

.PHONY: run-cli
run-cli:
	@echo "+ $@"
	$(GO) run cmd/cli/main.go
