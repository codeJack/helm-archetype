DIST := $(CURDIR)/_dist
HELM_PLUGINS = $(shell helm env | grep HELM_PLUGINS | sed -e 's/HELM_PLUGINS=\(.*\)/\1/g')
HELM_ARCHETYPE_DIR = ${HELM_PLUGINS}/helm-archetype
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)

.PHONY: all
all: deps build

.PHONY: clean
clean:
	go clean

.PHONY: deps
deps:
	go mod download
	go mod vendor

.PHONY: build
build:
	go build -o helm-archetype

.PHONY: test
test:
	go test ./...

.PHONY: install
install: deps build
	mkdir -p $(HELM_ARCHETYPE_DIR)
	cp helm-archetype $(HELM_ARCHETYPE_DIR)
	cp plugin.yaml $(HELM_ARCHETYPE_DIR)

.PHONY: deliver
deliver:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	tar -zcvf $(DIST)/helm-archetype-linux-$(VERSION).tgz helm-archetype
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
	tar -zcvf $(DIST)/helm-archetype-macos-$(VERSION).tgz helm-archetype
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
	tar -zcvf $(DIST)/helm-archetype-windows-$(VERSION).tgz helm-archetype.exe