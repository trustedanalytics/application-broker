#!/bin/bash

export CF_API="http://api.54.68.64.168.xip.io"
export CF_SRC="/Users/markchma/Code/spring-hello-env"
export CF_DEP="postgresql93|free,consul|free"
export CF_CAT="../catalog.json"

# for runtime you have to set this to
# export CF_CAT="./catalog.json"
