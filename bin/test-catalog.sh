#!/bin/bash

curl http://127.0.0.1:9999/v2/catalog \
    -H "Content-Type: application/json" \
    -X GET \
    -v
