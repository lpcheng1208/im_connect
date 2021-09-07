#!/bin/bash
.PHONY: sync

envT := prod
sync:
ifeq ($(envT), prod)
	sudo lsyncd scripts/sync.lua
else
	sudo lsyncd scripts/sync_test.lua
endif

default: clean

