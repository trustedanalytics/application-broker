#!/bin/bash

echo "==== Deregistering from catalog by sending DELETE on /v2/catalog ===="
serviceGuid=`cat service.guid`

status_code=`curl -sL $BROKER_ADDRESS/v2/catalog/$serviceGuid -X DELETE    \
    -u $AUTH_USER:$AUTH_PASS                                  \
    -H "Content-Type: application/json"      \
    -w "%{http_code}\\n"                     \
    -o deregister.log`

if [ "$status_code" == "204" ]
then
    echo "SUCCESS!"
else
    echo "FAILED! You have instances associated with this service probably!"
fi
