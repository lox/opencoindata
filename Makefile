DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+changes" || true)

all: deps build

build:
	@echo $(DEPS) 
	@mkdir -p bin/
	go get ./...
	go build -ldflags "-X main.GitCommit '$(GIT_COMMIT)$(GIT_DIRTY)'" -o bin/opencoindata

cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html

deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d

update:
	go get -u -v
	echo $(DEPS) | xargs -n1 go get -d -u 

test: deps 
	go list ./... | xargs -n1 go test

.PNONY: all cov deps integ test