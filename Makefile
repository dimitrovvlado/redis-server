# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)
REDIS := /usr/local/opt/redis/bin/redis-server

.PHONY: run-redis
run-redis:
	@echo "+ $@"
	$(REDIS) ./redis.conf
