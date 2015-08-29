#!/bin/bash

#Script that registers application in catalog of the Application Broker.

# Usage: ./register.sh brokerAddress basicAuthUser basicAuthPass nameOfAppToRegister marketName "marketDescription" marketIcon

# brokerAddress                  - address of the Application Broker
# basicAuthUser basicAuthPass    - credentials that broker API is secured with
# nameOfAppToRegister            - name of existing reference app that will be treated as service
# marketName                     - name of the offering in marketplace that will be created
# marketDescription              - explanation of what application being registered provides
# marketIcon                     - path to icon that will be placed in marketplace (or empty "")
#Caution!
#This script assumes that you are targeted to particular org and space with CF CLI

brokerAddress=$1
echo "Address of application broker: " $brokerAddress

basicAuthUser=$2

basicAuthPass=$3

appName=$4
echo "App name to register: " $appName

marketName=$5
echo "Name of service to register in marketplace: " $marketName

marketDesc=$6
echo "Application description in marketplace: " $marketDesc

marketIcon=$7
echo "Application icon in marketplace: " $marketIcon

if [ "$marketIcon" != "" ]
then
	#get file
	extension="${marketIcon##*.}"
	base64_encoded="$( base64 $marketIcon | tr -d '\n' )"
	metadata=', "metadata" : {"imageUrl":"data:image/'"$extension"';base64,'"$base64_encoded"'"}'
fi

applicationGuid=`cf app $appName --guid`
echo $applicationGuid

status_code=`curl -sL $brokerAddress/v2/catalog -X POST      \
    -u $basicAuthUser:$basicAuthPass                         \
    -H "Content-Type: application/json"                      \
	-d '{
            "app" : {"metadata" : {"guid" : "'$applicationGuid'"}},
            "description" : "'"${marketDesc}"'",
            "name" : "'$marketName'"
			'"$metadata"'
            }'			\
    -w "%{http_code}\\n"			\
	-o demoRegister.log`


if [ "$status_code" == "201" ] ; then
    echo "Application registered!"
elif [ "$status_code" == "409" ] ; then
    echo "Application already registered!"
else
    echo "FAILED!"
    exit 1
fi

cf enable-service-access $marketName
