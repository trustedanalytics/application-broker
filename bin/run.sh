#!/bin/bash

# do we have CF
if which cf >/dev/null; then
    echo "cf already in path"
else
    echo "downloading..."
    cd ./bin
    wget -O cf https://s3.amazonaws.com/go-cli/builds/cf-linux-amd64
    chmod u+x
    DIR=$(pwd)
    export PATH=$DIR/cf:$PATH
fi

./app-launching-service-broker
