#!/bin/bash

export CF_API=${CF_API:-"https://api.gotapaas.com"}
export CF_SRC=${CF_SRC:-"$PWD/apps/cf-env"}
export CF_DEP=${CF_DEP:-"consul|free,nats|free"}
export CF_CATALOG_PATH=${CF_CATALOG_PATH:-"$PWD/catalog.json"}
export CF_USER=${CF_USER:-"admin"}
export CLIENT_ID=${CLIENT_ID:-"admin"}
export CLIENT_SECRET=${CLIENT_SECRET:-"admin"}
export ROOT_DOMAIN=${ROOT_DOMAIN:-gotapaas.com}
export REDIRECT_URL=${REDIRECT_URL:-"https://my-client.$ROOT_DOMAIN"}
export AUTH_URL=${AUTH_URL:-"https://login.$ROOT_DOMAIN/oauth/authorize"}
export TOKEN_URL=${TOKEN_URL:-"https://uaa.$ROOT_DOMAIN/oauth/token"}


if [[ "${CF_PASS}X" == "X" ]]; then
  echo "ERROR: set \$CF_PASS for the $CF_USER password"
  exit 1
fi

if [[ "${CF_API_SKIP_SSL_VALID}X" == "X" ]]; then
  echo "If necessary, set \$CF_API_SKIP_SSL_VALID to skip SSL verification"
fi

# for runtime you have to set this to
# export CF_CATALOG_PATH="./catalog.json"
