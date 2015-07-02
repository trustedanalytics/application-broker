#!/bin/bash
#
# Copyright (c) 2015 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#


export CF_API=${CF_API:-"https://api.gotapaas.com"}
export CF_SRC=${CF_SRC:-"$PWD/apps/cf-env"}
export CF_DEP=${CF_DEP:-"consul|free,nats|free"}
export CF_SETUP_PATH=${CF_SETUP_PATH:-"$PWD/apps/cf-env/setup.sh"}
export CF_CATALOG_PATH=${CF_CATALOG_PATH:-"$PWD/catalog.json"}
export CF_USER=${CF_USER:-"admin"}
export UI=${UI:-"true"}
export ROOT_DOMAIN=${ROOT_DOMAIN:-gotapaas.com}
export REDIRECT_URL=${REDIRECT_URL:-"http://localhost:9999/oauth2callback"}
export AUTH_URL=${AUTH_URL:-"https://login.$ROOT_DOMAIN/oauth/authorize"}
export TOKEN_URL=${TOKEN_URL:-"https://uaa.$ROOT_DOMAIN/oauth/token"}
export API_URL=${API_URL:-"https://api.$ROOT_DOMAIN"}
export DASHBOARD_URL=${DASHBOARD_URL:-"http://foobar.com/instances/"}


if [[ "${CF_PASS}X" == "X" ]]; then
  echo "ERROR: set \$CF_PASS for the $CF_USER password"
  exit 1
fi

if [[ "${CF_API_SKIP_SSL_VALID}X" == "X" ]]; then
  echo "If necessary, set \$CF_API_SKIP_SSL_VALID to skip SSL verification"
fi

# for runtime you have to set this to
# export CF_CATALOG_PATH="./catalog.json"
