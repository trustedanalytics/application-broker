#!/bin/bash

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
