#!/bin/bash

./app-launching-service-broker -h="127.0.0.1" \
                               -p=9999 \
                               -u=$CF_USER \
                               -s=$CF_PASS \
                               --api="http://api.54.68.64.168.xip.io" \
                               --cfu=$CF_USER \
                               --cfp=$CF_PASS \
                               --src="/Users/markchma/Code/spring-hello-env" \
                               --dep="postgresql93|free,consul|free" \
                               --cat="./catalog.json"

