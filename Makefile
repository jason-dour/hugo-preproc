RM := rm -f

MAKEFILE = $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))

MODULE = $(shell awk '/module/{print $$2}' go.mod)
BASENAME = $(lastword $(subst /, , $(MODULE)))
VERSION = $(shell cat VERSION)
LDFLAGS = "-X $(MODULE)/cmd.version=$(VERSION) -X $(MODULE)/cmd.basename=$(BASENAME)"

all: build

.PHONY: all

build:
	@go build -ldflags $(LDFLAGS) 
.PHONY: build

windows:
	@env GOOS=windows GOARCH=386 go build -ldflags $(LDFLAGS) -o $(BASENAME)_v$(VERSION)_win32_x32.exe
	@env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $(BASENAME)_v$(VERSION)_win32_x64.exe
.PHONY: windows

linux:
	@env GOOS=linux GOARCH=386 go build -ldflags $(LDFLAGS) -o $(BASENAME)_v$(VERSION)_linux_x32
	@env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $(BASENAME)_v$(VERSION)_linux_x64
.PHONY: linux

macos:
	@env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $(BASENAME)_v$(VERSION)_darwin_x64
.PHONY: macos

release: windows linux macos
.PHONY: release

clean:
	@$(RM) $(BASENAME) $(BASENAME).v*
.PHONY: clean
