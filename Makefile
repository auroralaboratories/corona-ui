all: fmt deps build

fmt:
	gofmt -w ./..

deps:
	go get .

build:
	pkg-config --libs 'webkit2gtk-4.0 >= 2.8'
	@which go-bindata > /dev/null 2>&1 || go get github.com/jteeuwen/go-bindata/...
	which go-bindata
	@go-bindata --pkg util --prefix embed embed
	@mv bindata.go util/
	CC=gcc-4.9 go build -o bin/`basename ${PWD}`
