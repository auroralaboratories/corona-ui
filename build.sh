#!/usr/bin/env bash

PROJECT="${PROJECT:-$(basename $PWD)}"
ORG_PATH="github.com/auroralaboratories"
REPO_PATH="${ORG_PATH}/${PROJECT}"

export GOPATH=${PWD}/gopath
export PATH="$GOPATH/bin:$PATH"

rm -f $GOPATH/src/${REPO_PATH}
mkdir -p $GOPATH/src/${ORG_PATH}
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

eval $(go env)

if [ -s DEPENDENCIES ]; then
  echo 'Processing dependencies...'
  for f in $(find . -name DEPENDENCIES); do
    for d in $(cat ${f} | sort | uniq); do
      if [ "${d}" == "github.com/conformal/gotk3/gtk" ]; then
        go get -v -tags gtk_3_10 $d
      else
        go get $d
      fi
    done
  done
fi

# apply patches
patch -r'-' -N -p0 < patches/all.patch

# build the go-bindata tool
# echo '-> building go-bindata utility'
# cd gopath/src/github.com/jteeuwen/go-bindata/go-bindata
# go build
# cd - > /dev/null

# echo '-> building go-bindata-assetfs utility'
# cd gopath/src/github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs
# go build
# cd - > /dev/null

# export PATH="$PWD/gopath/src/github.com/jteeuwen/go-bindata/go-bindata:$PWD/gopath/src/github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs:$PATH"
# echo 'Embedding static assets from ./public/'
# go-bindata-assetfs --pkg util $(find public -type d | tr "\n" " ")
# sed -i 's/func assetFS()/func AssetFS()/' bindata_assetfs.go
# mv bindata_assetfs.go util/

# set flags
[ "$DEBUG" == 'true' ] || GOFLAGS="-ldflags '-s'"

# build it!
echo "Building..."
go build -tags gtk_3_10 -a $GOFLAGS -o bin/${PROJECT} ${REPO_PATH}/


# vendor the dependencies
echo 'Vendoring...'
# remove all .git directories except the local projects (that would be bad :)
find gopath -type d | grep -v "${REPO_PATH}" | grep -v ^\./\.git$ | grep \.git$ | xargs rm -rf