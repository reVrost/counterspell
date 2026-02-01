SHELL := /bin/sh

CS_PREFIX ?= counterspell
CS_REMOTE ?= counterspell-remote
CS_BRANCH ?= main

.PHONY: subtree-pull subtree-push

# Pulls latest changes from Counterspell public repo into the monorepo
subtree-pull:
	git subtree pull --prefix=$(CS_PREFIX) $(CS_REMOTE) $(CS_BRANCH) --squash

# Pushes monorepo Counterspell changes to the public repo
subtree-push:
	git subtree push --prefix=$(CS_PREFIX) $(CS_REMOTE) $(CS_BRANCH)
