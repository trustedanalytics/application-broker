#!/bin/bash

curl http://localhost:8888/v2/service_instances/640A1E39-D5A4-408D-85E5-72A44A383425 \
    -H "Content-Type: application/json" \
    -X DELETE \
    -v
