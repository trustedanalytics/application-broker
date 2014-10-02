#!/bin/bash

curl http://127.0.0.1:9999/v2/catalog \
    -H "Content-Type: application/json" \
    -H "X-Broker-Api-Version: 2.3" \
    -H "Authorization: bearer dGVzdC11c2VyOnRlc3QtcGFzc3dvcmQ=" \
    -X GET \
    -v
