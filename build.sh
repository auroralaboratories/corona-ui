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
      if [ "${d}" == "github.com/gotk3/gotk3/gtk" -o "${d}" == "github.com/auroralaboratories/go-webkit2/webkit2" ]; then
        go get -v -tags gtk_3_10 $d
      else
        go get $d
      fi
    done
  done
fi

# apply patches
rsync -rv ./patches/ ./gopath/src/

# set flags
[ "$DEBUG" == 'true' ] || GOFLAGS="-ldflags '-s'"

# build it!
echo "Building..."
go build -tags gtk_3_10 -a $GOFLAGS -o bin/${PROJECT} ${REPO_PATH}/


# vendor the dependencies
echo 'Vendoring...'
# remove all .git directories except the local projects (that would be bad :)
find gopath -type d | grep -v "${REPO_PATH}" | grep -v ^\./\.git$ | grep \.git$ | xargs rm -rf