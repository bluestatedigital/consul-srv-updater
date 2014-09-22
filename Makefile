NAME=consul-srv-updater
BIN=.godeps/bin

GPM=$(BIN)/gpm
GVP=$(BIN)/gvp

SOURCES=$(shell go list -f '{{range .GoFiles}}{{.}} {{end}}' ./... )

.PHONY: all init build tools release clean

all: build

$(BIN):
	mkdir -p $(BIN)

$(GPM): $(BIN)
	curl -s -L -o $@ https://github.com/pote/gpm/raw/v1.2.3/bin/gpm
	chmod +x $@

$(GVP): $(BIN)
	curl -s -L -o $@ https://github.com/pote/gvp/raw/v0.1.0/bin/gvp
	chmod +x $@

tools: $(GPM) $(GVP)

.godeps/src: tools
	$(GVP) init
	$(GVP) in $(GPM) install

init: .godeps/src
	mkdir -p stage

build: stage/$(NAME)

stage/$(NAME): init $(SOURCES)
	$(GVP) in go build -v -o $@ ./...

clean:
	rm -rf stage release .godeps
