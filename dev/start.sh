#!/usr/bin/env sh

set -x

if ! mongod --version > /dev/null 2>&1
then
    echo "mongodb is required, please install it first"
    exit 1
fi

cd "$(dirname "${0}")/.."


export PATH=$PATH:${GOPATH:-$HOME/go}/bin

if ! goreman help > /dev/null 2>&1
then
    echo "install goreman..."
    if ! go get github.com/mattn/goreman > /dev/null 2>&1
    then
        echo "goreman install failed"
        exit 1
    fi
fi

echo "start to build"

services="account gits index project repository syntect api"

for service in ${services}
do
    echo "building ${service}"
    if ! go build -o build/${service} -mod=vendor github.com/lt90s/rfschub-server/${service}/cmd/server
    then
        echo "building ${service} failed"
        exit 1
    fi
done

exec goreman -f dev/Procfile start