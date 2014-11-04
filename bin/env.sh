#!/bin/bash

export CF_API=${CF_API:-"https://api.gotapaas.com"}
export CF_SRC=${CF_SRC:-"/Users/drnic/Projects/cloudfoundry/apps/spring-music"}
export CF_DEP=${CF_DEP:-"postgresql93|free,consul|free"}
export CF_CATALOG_PATH=${CF_CATALOG_PATH:-"./catalog.json"}
export CF_USER=${CF_USER:-"admin"}

if [[ "${CF_PASS}X" == "X" ]]; then
  echo "Remember to set \$CF_PASS for the $CF_USER password"
fi

if [[ "${CF_API_SKIP_SSL_VALID}X" == "X" ]]; then
  echo "If necessary, set \$CF_API_SKIP_SSL_VALID to skip SSL verification"
fi

# for runtime you have to set this to
# export CF_CATALOG_PATH="./catalog.json"
