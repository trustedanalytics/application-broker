#!/bin/bash

export CF_API="https://api.gotapaas.com"
export CF_SRC="/Users/markchma/Code/spring-hello-env"
export CF_DEP="postgresql93|free,consul|free"
export CF_CAT="../catalog.json"

# for runtime you have to set this to
# export CF_CAT="./catalog.json"
