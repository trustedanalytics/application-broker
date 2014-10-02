#!/bin/bash

curl http://127.0.0.1:9999/v2/service_instances/640A1E39-D5A4-408D-85E5-72A44A383425/service_bindings/642E57F0-492D-4BB8-8C98-CAF2651BF523 \
    -H "Content-Type: application/json" \
    -H "X-Broker-Api-Version: 2.3" \
    -H "Authorization: bearer dGVzdC11c2VyOnRlc3QtcGFzc3dvcmQ=" \
    -X PUT \
    -v \
    -d '{
            "plan_id": "29140B3F-0E69-4C7E-8A35-7AB2805AC4AC",
            "service_id": "45E600DA-D081-4188-85F2-64767BE0E836",
            "app_guid": "642E57F0-492D-4BB8-8C98-CAF2651BF523"
    }'