#!/bin/bash

curl http://127.0.0.1:9999/v2/service_instances/640A1E39-D5A4-408D-85E5-72A44A383425 \
    -H "Content-Type: application/json" \
    -H "X-Broker-Api-Version: 2.3" \
    -H "Authorization: Basic dGVzdC11c2VyOnRlc3QtcGFzc3dvcmQ=" \
    -X DELETE \
    -v
