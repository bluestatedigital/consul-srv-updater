NAME=consul-srv-updater
BIN=.godeps/bin

GPM=$(BIN)/gpm
GVP=$(BIN)/gvp

SOURCES=$(shell go list -f '{{range .GoFiles}}{{.}} {{end}}' ./... )

.PHONY: all init build tools clean

all: build

$(BIN) stage:
	mkdir -p $@

$(GPM): $(BIN)
	curl -s -L -o $@ https://github.com/pote/gpm/raw/v1.2.3/bin/gpm
	chmod +x $@

$(GVP): $(BIN)
	curl -s -L -o $@ https://github.com/pote/gvp/raw/v0.1.0/bin/gvp
	chmod +x $@

tools: $(GPM) $(GVP)

.godeps: $(GVP)
	$(GVP) init

## can't use "tools" as a target because it's not a real file
.godeps/.gpm_installed: .godeps $(GPM) $(GVP) Godeps
	$(GVP) in $(GPM) install
	touch $@

init: .godeps/.gpm_installed stage

build: stage/$(NAME)

stage/$(NAME): init $(SOURCES)
	$(GVP) in go build -v -o $@ ./...

clean:
	rm -rf stage .godeps
