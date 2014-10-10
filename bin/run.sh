#!/bin/bash

# do we have CF
if which cf >/dev/null; then
    echo "cf already in path"
else
    cd ./bin
    echo "downloading..."
    wget -O cf-cli.tgz https://cli.run.pivotal.io/stable?release=linux64-binary&version=6.6.1
    echo "uncompressing..."
    tar -zxvf cf-cli.tgz
    echo "making executable..."
    su -c "chmod +x cf"
    echo "adding to PATH..."
    DIR=$(pwd)
    export PATH=$DIR/cf:$PATH
fi

./app-launching-service-broker
