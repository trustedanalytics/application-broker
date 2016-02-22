#!/bin/bash

export GOPATH=$GOPATH:$(pwd)/Godeps/_workspace/
go get -u github.com/go-swagger/go-swagger/cmd/swagger
mkdir ./public
swagger generate spec -o ./public/swagger.json
