#!/bin/bash

source ./bin/env.sh

export CF_CATALOG_PATH="$PWD/catalog.json"

go test ./... --short
