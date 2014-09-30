#!/bin/bash

./app-launching-service-broker -h=127.0.0.1 \
                               -p=9999 \
                               -u=user \
                               -s=pswd \
                               -d=true \
                               -src="./spring-music" \
                               -dep="postgresql93|free,consul|free"

