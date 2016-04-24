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
	@which go-bindata > /dev/null 2>&1 || go get github.com/jteeuwen/go-bindata/...
	which go-bindata
	@go-bindata --pkg util --prefix embed embed
	@mv bindata.go util/
	CC=gcc-4.9 go build -o bin/`basename ${PWD}`
