#!/bin/bash

source ./bin/env.sh

# override this one
export CF_CATALOG_PATH="./catalog.json"

./app-launching-service-broker
