#!/bin/bash

function show_help {
    echo '
Script that removes application from catalog of the Application Broker.

Usage: ./deregister.sh -b <brokerAddress> -u <basicAuthUser> -p <basicAuthPass> -s <serviceID>

-b brokerAddress          - address of the Application Broker
-u basicAuthUser          - broker credentials - user
-p basicAuthPass          - broker credentials - password
-n serviceName            - Name of the service offering that will be removed

Caution!
This script assumes that you are targeted to particular org and space with CF CLI
'
}

while getopts "b:u:p:n:h" optname; do
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
        "n")
            serviceName=$OPTARG
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

echo "Address of application broker: " $brokerAddress
echo "Name of the service offering to remove: " $serviceName
serviceID=`cf curl /v2/services | jq ".resources[] | select(.entity.label==\"$serviceName\") | .entity.unique_id" | tr -d '"'`

echo "ID of the service offering to remove: " $serviceID
if [ "$serviceID" == "" ]; then
    echo "Could not conclude ID of service"
    exit 1
fi

statusCode=`curl -sL $brokerAddress/v2/catalog/$serviceID -X DELETE  \
    -u $basicAuthUser:$basicAuthPass                         \
    -H "Content-Type: application/json"                      \
    -w "%{http_code}\\n"                                     \
    -o deregister.log`

if [ "$statusCode" == "204" ] ; then
    echo "Service deregistered!"
elif [ "$statusCode" == "404" ] ; then
    echo "Service does not exist already!"
else
    echo "FAILED!"
    exit 1
fi
