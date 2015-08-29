#!/bin/bash

echo "==== setting cf to point to proper environment ===="
cf api $CF_API
cf auth $CF_USER $CF_PASS
cf target -o $CF_ORG -s $CF_SPACE

echo "==== deleting space $CF_SPACE from $CF_ORG to clean all apps ===="
cf delete-space $CF_SPACE -f

echo "==== recreating $CF_SPACE ===="
cf create-space $CF_SPACE

echo "==== retargeting ===="
cf target -o $CF_ORG -s $CF_SPACE

echo "==== push sample app (it will be reference app) ===="
cd sampleApp
referenceAppName=referenceApp-`cat /dev/urandom | tr -cd 'a-f0-9' | head -c 5`
cf push $referenceAppName
export sampleAppGuid=`cf app $referenceAppName --guid`
cd ..

echo "==== saving sample app guid in sampleApp.guid file ===="
echo $sampleAppGuid > sampleApp.guid

mongoGuid=`cat /dev/urandom | tr -cd 'a-f0-9' | head -c 10`
memcachedGuid=`cat /dev/urandom | tr -cd 'a-f0-9' | head -c 10`
cf cs mongodb26 free $mongoGuid
cf cs memcached14 128Mb $memcachedGuid
cf bs $referenceAppName $mongoGuid
cf bs $referenceAppName $memcachedGuid
cf se $referenceAppName TEST_ENV_NAME TEST_ENV_VAL
