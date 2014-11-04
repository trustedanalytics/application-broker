#!/bin/bash

source ./bin/env.sh

export CF_CATALOG_PATH="./catalog.json"

go test ./... --short
