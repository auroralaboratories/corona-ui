all: fmt build

deps:
	go list github.com/Masterminds/glide
	glide install

fmt:
	gofmt -w .

build:
	rsync -rv ./patches/ ./vendor/
	go build -o bin/`basename ${PWD}`
