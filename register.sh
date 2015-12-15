#!/bin/bash

function show_help {
    echo '
Script that registers application in catalog of the Application Broker.

Usage: ./register.sh  -b <brokerAddress> -u <basicAuthUser> -p <basicAuthPass> -a <nameOfAppToRegister> -n <marketName> -d <marketDescription> [-s <marketDisplayName] [-i <marketIcon>]

-b brokerAddress          - address of the Application Broker
-u basicAuthUser          - broker credentials - user
-p basicAuthPass          - broker credentials - password
-a nameOfAppToRegister    - name of existing reference app that will be treated as service
-n marketName             - name of the service offering that will be created
-s displayName (optiona)  - display name that will be visible in the marketplace
-d marketDescription      - explanation of what application being registered provides
-i marketIcon (optional)  - path to icon that will be placed in marketplace (or empty "")

Caution!
This script assumes that you are targeted to particular org and space with CF CLI
'
}

while getopts "b:u:p:a:n:s:d:i:h" optname; do
    case "$optname" in
        "b")
            brokerAddress=$OPTARG
            ;;
        "u")
            basicAuthUser=$OPTARG
            ;;
        "p")
            basicAuthPass=$OPTARG
            ;;
        "a")
            appName=$OPTARG
            ;;
        "n")
            marketName=$OPTARG
            ;;
        "s")
            displayName=$OPTARG
            ;;
        "d")
            marketDesc=$OPTARG
            ;;
        "i")
            marketIcon=$OPTARG
            ;;
        "h")
            show_help
            exit
            ;;
        "?")
            echo "Unknown option $OPTARG"
            exit
            ;;
        ":")
            echo "No argument value for option $OPTARG"
            ;;
        *)
            # Should not occur
            echo "Unknown error while processing options"
            ;;
    esac
done

if [ -n "$displayName" ]; then
    displayName=$marketName
fi

echo "Address of application broker: " $brokerAddress
echo "App name to register: " $appName
echo "Name of the service offering: " $marketName
echo "Display name (name that will appear in marketplace): " $displayName
echo "Application description in marketplace: " $marketDesc
echo "Application icon in marketplace: " $marketIcon

if [[ "$marketIcon" ]] && [[ "$marketIcon" != "data:image"* ]]
then
	#get file
	extension="${marketIcon##*.}"
	base64_encoded="$( base64 $marketIcon | tr -d '\n' )"
	marketIcon='data:image/'"$extension"';base64,'"$base64_encoded"
fi

metadata='{
    "displayName": "'$displayName'",
    "imageUrl": "'$marketIcon'"
}'

applicationGuid=`cf app $appName --guid`
echo $applicationGuid

status_code=`curl -sL $brokerAddress/v2/catalog -X POST      \
    -u $basicAuthUser:$basicAuthPass                         \
    -H "Content-Type: application/json"                      \
	-d '{
            "app" : {"metadata" : {"guid" : "'$applicationGuid'"}},
            "description" : "'"${marketDesc}"'",
            "name" : "'$marketName'",
			"metadata" : '"$metadata"'
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
