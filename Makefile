DOCKER_BIN=$(shell which docker)

all:
	go build -o metathings-module-joker cmd/module/main.go
	go build -o joker cmd/joker/*.go

protos:
	$(DOCKER_BIN) run --rm -v $(PWD):/go/src/github.com/nayotta/metathings-component-joker nayotta/metathings-development /usr/bin/make -C /go/src/github.com/nayotta/metathings-component-joker/proto
