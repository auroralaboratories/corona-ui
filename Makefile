.PHONY: test deps fmt build

all: fmt deps build

fmt:
	gofmt -w ./..

deps:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	go generate -x
	go get .

build:
	pkg-config --libs 'webkit2gtk-4.0 >= 2.8'
	CC=gcc-4.9 go build -o bin/`basename ${PWD}`
