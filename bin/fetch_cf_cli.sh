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


URL=https://s3.amazonaws.com/go-cli/builds/cf-linux-amd64

if [[ "$(which wget)X" != "X" ]]; then
  wget -O ./bin/cf $URL
elif [[ "$(which curl)X" != "X" ]]; then
  curl -L -o apps/cf $URL
else
  echo "Install wget or curl"
  exit 1
fi

chmod +x ./bin/cf
