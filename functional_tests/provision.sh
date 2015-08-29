#!/bin/bash

echo "==== provisioning new service instance by sending PUT on /v2/service_instances/ ===="

instanceGuid=`uuidgen`
echo $instanceGuid > instance.guid
instanceName=`cat /dev/urandom | tr -cd 'a-f0-9' | head -c 10`
echo $instanceName > instance.name

orgGuid=`cf org $CF_ORG --guid`
spaceGuid=`cf space $CF_SPACE --guid`
planGuid=`cat plan.guid`
serviceGuid=`cat service.guid`

status_code=`curl -sL $BROKER_ADDRESS/v2/service_instances/$instanceGuid -X PUT      \
    -u $AUTH_USER:$AUTH_PASS                                                         \
    -H "Content-Type: application/json"      \
    -d '{
            "organization_guid" : "'$orgGuid'",
            "plan_id"           : "'$planGuid'",
            "service_id"        : "'$serviceGuid'",
            "space_guid"        : "'$spaceGuid'",
            "parameters"        : {"name" : "'$instanceName'"}
        }'                                   \
    -w "%{http_code}\\n"                     \
    -o provision.log`

env_vars=`cf env $instanceName`

if [ "$status_code" == "201" ] && [[ $env_vars == *"TEST_ENV_NAME"* ]] && [[ $env_vars == *"TEST_ENV_VAL"* ]]
then
    echo "SUCCESS!"
else
    echo "FAILED!"
fi
