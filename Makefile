.DEFAULT_GOAL := help

ADMIN_HOST ?= 127.0.0.1
ADMIN_PORT ?= 3000
BRAIN_API_ORIGIN ?= http://127.0.0.1:18082
WORKERS ?= 4
WORKER_SLOTS ?= 1

.PHONY: help setup dev server admin

ifneq ($(filter dev,$(MAKECMDGOALS)),)
ifneq ($(filter server admin,$(MAKECMDGOALS)),)
ifneq ($(words $(filter server admin,$(MAKECMDGOALS))),1)
$(error Usage: make dev, make dev server, or make dev admin)
endif
endif
endif

help:
	@printf '%s\n' 'make setup       Install server and admin Node dependencies' 'make dev         Start Brain, admin and four local Neurons (WORKERS=4)' 'make dev server  Start Brain and local Neurons without admin' 'make dev admin   Start only the Next.js admin'

setup:
	npm ci
	npm --prefix admin-nextjs ci

dev:
ifeq ($(filter server admin,$(MAKECMDGOALS)),)
	ADMIN_HOST="$(ADMIN_HOST)" ADMIN_PORT="$(ADMIN_PORT)" BRAIN_API_ORIGIN="$(BRAIN_API_ORIGIN)" BRAIN_DEV_WORKERS="$(WORKERS)" BRAIN_DEV_SLOTS="$(WORKER_SLOTS)" python3 scripts/dev-stack.py
else
	@:
endif

server:
	BRAIN_DEV_WORKERS="$(WORKERS)" BRAIN_DEV_SLOTS="$(WORKER_SLOTS)" bash scripts/start-local.sh

admin:
	BRAIN_API_ORIGIN="$(BRAIN_API_ORIGIN)" npm --prefix admin-nextjs run dev -- --hostname "$(ADMIN_HOST)" --port "$(ADMIN_PORT)"
