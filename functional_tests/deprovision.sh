#!/bin/bash

echo "==== deprovisioning new service instance by sending DELETE on /v2/service_instances/ ===="
instanceGuid=`cat instance.guid`
instanceName=`cat instance.name`

status_code=`curl -sL $BROKER_ADDRESS/v2/service_instances/$instanceGuid -X DELETE      \
    -u $AUTH_USER:$AUTH_PASS                 \
    -H "Content-Type: application/json"      \
    -w "%{http_code}\\n"                     \
    -o /dev/null`

servicesCount=`cf s | grep "$instanceName" | wc -l`
routesCount=`cf r | grep "$instanceName" | wc -l`

echo "Status code: $status_code"
echo "Services left: $servicesCount"
echo "Routes left: $routesCount"

if [ "$status_code" == "200" ] && [ "$servicesCount" == "0" ] && [ "$routesCount" == "0" ]
then
    echo "SUCCESS!"
else
    echo "FAILED!"
fi
