#!/bin/bash

source ./env.sh

# override this one
export CF_CAT="./catalog.json"

./app-launching-service-broker
