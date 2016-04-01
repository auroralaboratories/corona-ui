all: fmt build

deps:
	go list github.com/Masterminds/glide
	glide install

fmt:
	gofmt -w .

build:
	go build -tags gtk_3_10 -o bin/`basename ${PWD}`
