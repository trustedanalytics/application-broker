#!/bin/bash

curl http://localhost:8888/v2/catalog \
    -H "Content-Type: application/json" \
    -X GET \
    -v
