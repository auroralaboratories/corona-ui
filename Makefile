.PHONY: test deps fmt build

all: fmt deps build

fmt:
	gofmt -w ./..

deps:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	go generate -x
	go get .

build:
	pkg-config --libs 'webkit2gtk-4.0 >= 2.8'
	go build -i -o bin/`basename ${PWD}`
