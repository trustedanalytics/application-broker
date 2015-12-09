#!/bin/bash

echo "==== Cleaning all services registered in catalog ===="
mongo application-broker --eval "db.dropDatabase()"


echo "==== Registering in catalog by sending POST on /v2/catalog ===="
sampleAppGuid=`cat sampleApp.guid`
planGuid=`uuidgen`
echo $planGuid > plan.guid
serviceGuid=`uuidgen`
echo $serviceGuid > service.guid
inserted_data=`curl -sL $BROKER_ADDRESS/v2/catalog -X POST      \
    -u $AUTH_USER:$AUTH_PASS                                  \
    -H "Content-Type: application/json"      \
    -d '{
            "app" : {
                        "metadata" : {"guid" : "'$sampleAppGuid'"}
                    },
            "id" : "'$serviceGuid'",
            "plans" : [{"id" : "'$planGuid'"}],
            "description" : "as simple as possible",
            "name" : "sampleAppTEST"
        }'`

update_data=$(echo $inserted_data | sed 's/sampleAppTEST/sampleAppService/g')

echo $update_data

echo "==== Updating data ===="

status_code=`curl -sL $BROKER_ADDRESS/v2/catalog/$serviceGuid -X PUT      \
    -u $AUTH_USER:$AUTH_PASS                                  \
    -H "Content-Type: application/json"                      \
	-d ''"$update_data"''					\
    -w "%{http_code}\\n"                     \
    -o update.log`

if [ "$status_code" == "200" ]
then
    echo "SUCCESS!"
else
    echo "FAILED!"
fi

