.DEFAULT_GOAL := help

ADMIN_HOST ?= 127.0.0.1
ADMIN_PORT ?= 3000
BRAIN_API_ORIGIN ?= http://127.0.0.1:18082

.PHONY: help setup dev server admin

ifneq ($(filter dev,$(MAKECMDGOALS)),)
ifneq ($(words $(filter server admin,$(MAKECMDGOALS))),1)
$(error Usage: make dev server or make dev admin)
endif
endif

help:
	@printf '%s\n' 'make setup       Install server and admin Node dependencies' 'make dev server  Start the local Go server' 'make dev admin   Start the Next.js admin'

setup:
	npm ci
	npm --prefix admin-nextjs ci

dev:
	@:

server:
	bash scripts/start-local.sh

admin:
	BRAIN_API_ORIGIN="$(BRAIN_API_ORIGIN)" npm --prefix admin-nextjs run dev -- --hostname "$(ADMIN_HOST)" --port "$(ADMIN_PORT)"
