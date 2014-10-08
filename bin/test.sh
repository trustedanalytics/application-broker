#!/bin/bash

source ./bin/env.sh

export CF_CAT="./catalog.json"

go test ./... --short
