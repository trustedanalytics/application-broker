#!/bin/bash

echo "==== creating new service binding by sending PUT on /v2/service_instances/:instance_id/service_bindings/:binding_id ===="
bindingGuid=`uuidgen`
echo $bindingGuid > binding.guid
instanceGuid=`cat instance.guid`

status_code=`curl -sL $BROKER_ADDRESS/v2/service_instances/$instanceGuid/service_bindings/$bindingGuid -X PUT      \
-u $AUTH_USER:$AUTH_PASS                                                         \
-H "Content-Type: application/json"                                              \
-d '{}'                                                                          \
-w "%{http_code}\\n"                                                             \
-o createBinding.log`

if [ "$status_code" == "200" ]
then
    echo "STATUS CODE OK!"
else
    echo "WRONG STATUS CODE!"
fi


#checking whether url was returned in credentials
set -o pipefail
responseUrl=`cat createBinding.log | python -mjson.tool | grep url | xargs | grep -oP 'url: \K[^"]+'`
if [ "$responseUrl" != "" ]
then
    echo "CREDENTIALS RETURNED - SUCCESS!"
    echo "URL returned: "$responseUrl
else
    echo "INCORRECT RETURN BODY"
fi
