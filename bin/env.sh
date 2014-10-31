#!/bin/bash

export CF_API=${CF_API:-"https://api.gotapaas.com"}
export CF_SRC=${CF_SRC:-"/Users/drnic/Projects/cloudfoundry/apps/spring-music"}
export CF_DEP=${CF_DEP:-"postgresql93|free,consul|free"}
export CF_CAT=${CF_CAT:-"./catalog.json"}

# for runtime you have to set this to
# export CF_CAT="./catalog.json"
