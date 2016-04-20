all: vendor fmt build

update:
	glide up --strip-vcs --update-vendored

vendor:
	go list github.com/Masterminds/glide
	glide install --strip-vcs --update-vendored

fmt:
	gofmt -w .

build:
	pkg-config --libs 'webkit2gtk-4.0 >= 2.8'
	rsync -rv ./patches/ ./vendor/
	CC=gcc-4.9 go build -o bin/`basename ${PWD}`
