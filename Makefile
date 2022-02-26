RM := rm -f

MAKEFILE = $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))

all: build

.PHONY: all

build:
	@go build

.PHONY: build

windows:
	@GOOS=windows
	@GOARCH=386
	@go build -o hugo-preproc.exe
	@zip -9 hugo-preproc.32bit.windows.zip hugo-preproc.exe
	@$(RM) hugo-preproc.exe
	@GOARCH=amd64
	@go build -o hugo-preproc.exe
	@zip -9 hugo-preproc.64bit.windows.zip hugo-preproc.exe
	@$(RM) hugo-preproc.exe
.PHONY: windows

linux:
	@GOOS=linux
	@GOARCH=386
	@go build -o hugo-preproc
	@tar -zcf hugo-preproc.32bit.linux.tar.gz hugo-preproc
	@$(RM) hugo-preproc
	@GOARCH=amd64
	@go build -o hugo-preproc
	@tar -zcf hugo-preproc.64bit.linux.tar.gz hugo-preproc
	@$(RM) hugo-preproc
.PHONY: linux

macos:
	@GOOS=darwin
	@GOARCH=amd64
	@go build -o hugo-preproc
	@tar -zcf hugo-preproc.64bit.macos.tar.gz hugo-preproc
	@$(RM) hugo-preproc
.PHONY: macos

release: windows linux macos
.PHONY: release

clean:
	@$(RM) *.zip *.tar.gz hugo-preproc
.PHONY: clean
