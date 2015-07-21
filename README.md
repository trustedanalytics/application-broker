App Launching Service Broker for Cloud Foundry
==============================================

A service broker to provision an application, including dependent service instances.

**This is a Proof of Contept and is subject to severe changes in the future.**

Usage
-----

Once the service broker is running as an app, registered as a service broker, and enabled for access by users, a user can use like any other service broker:

```
cf create-service some-service some-plan some-service-name
cf bind-service my-app some-service-name
cf restage my-app
```

The client application `my-app` would have a `$VCAP_SERVICE` service for `some-service`. The credentials would include the hostname, username and password to access the backend service application.

Behind the scenes, when `cf create-service` was invoked the CLI asked the Cloud Controller for the new service instance. The Cloud Controller asked the App Launching Service Broker for a new service instance. The App Launching Service Broker deploys a new backend application.

If the backend application requires any services for itself, then those too are created and bound to the new backend application.

The backend application and its own dependency services are created into the same organization and space being used by the end user.

![](app_launcher_workflow.jpg)

Development
-----------

To locally develop this service broker, you need to clone down an example application that will be deployed. Something small/fast to deploy will make your life better.

As an example, use the [spring-music](https://github.com/cloudfoundry-samples/spring-music) application. It is interesting as it supports a range of backend services.

Alternately, try the [cf-env](https://github.com/cloudfoundry-community/cf-env) application which has the sole purpose to display its environment variables.

```
cd path/to/apps
git clone https://github.com/cloudfoundry-community/cf-env
cd cf-env
bundle
export CF_SRC=$(pwd)
cd -
./bin/env.sh
```

This will create a set of environment variables used to configure the service broker:

```
$ env | grep CF
CF_API=https://api.<platform_domain>
CF_SRC=/users/myself/Projects/cloudfoundry/apps/cf-env
CF_DEP=postgresql93|free,consul|free
CF_CATALOG_PATH=./catalog.json
```

You now need to configure which admin user credentials the broker will use to communicate with Cloud Foundry:

```
export CF_USER=adminuser
export CF_PASS=adminpass
```

If you are using self-signed certificates, you may need to ignore SSL verification:

```
export CF_API_SKIP_SSL_VALID=true
```

Finally, load other default environment variables:

```
source bin/env.sh
```

You can now run the service broker locally via [gin](https://github.com/codegangsta/gin), which will automatically reload any file changes during development:

```
go get github.com/codegangsta/gin
gin
```

The broker, via `gin`, will be running on port 3000 by default.

```
$ curl http://localhost:3000/v2/catalog
{"services":[{"id":"XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX","name":"cf-env","description":"View environment variables","bindable":true,"tags":["demo","backend"],"plans":[{"id":"YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY","name":"simple","description":"Simple","free":true}]}]}
```

The output here matches the contents of the `./catalog.json` example file.

You can now even register your local app with a remote Cloud Foundry using [ngrok](https://ngrok.com/). Run the following in another terminal:

```
$ ngrok 3000
```

It will display a public URL for your local broker app.

Register your broker app URL:

```
$ cf create-service-broker cf-env-broker admin admin http://3f1c1555.ngrok.com
$ cf enable-service-access cf-env
```

You can now create service instances, which will deploy the local example app `cf-env`:

```
$ cf m
$ cf create-service cf-env simple cf-env-example
```

To clean up, delete services, then delete the service broker:

```
$ cf delete-service cf-env-example
$ cf delete-service-broker cf-env-broker
```

Testing
-------

```
git clone https://github.com/cloudfoundry-community/cf-env apps/cf-env
cd apps/cf-env
bundle
cd -

export CF_API=https://api.MYCLOUDFOUNDRY
export CF_PASS=<admin-password>
./bin/test.sh
```

Deploy to Cloud Foundry
-----------------------

The App Launching Service Broker is both *for* Cloud Foundry and can *run* on Cloud Foundry. This section shows how to deploy/run it on Cloud Foundry.

In this example, the broker will be deployed to launch the [cf-env](https://github.com/cloudfoundry-community/cf-env) sample app (which matches the sample `./catalog.json`):

```
export SERVICE=cf-env-launching
export APPNAME=$SERVICE-service-broker
git clone https://github.com/platform-data/app-launching-service-broker.git $APPNAME
cd $APPNAME
```

The broker will need access to the Cloud Foundry CLI within the Linux container it runs within. For this we need to download a version of the CLI that works with the target Cloud Foundry API:

```
./bin/fetch_cf_cli.sh
```

You would modify the `catalog.json` to document the application to be offered as a service. In this example, the included `catalog.json` corresponds to `cf-env`.

For the service ID and plan ID, you need unique UUIDs. Run the `uuid` command to generate different UUIDs and replace them into the `catalog.json`. Cloud Foundry will complain later if you try to register a service broker that uses the same UUIDs as existing brokers.

```
$ uuid
7d6f6d2a-6440-11e4-a6b5-6c4008a663f0
$ uuid
7dd52d9a-6440-11e4-b30c-6c4008a663f0
```

You need to embed the target application-as-a-service into the source code tree (in future the target application source will be fetched at runtime from remote blobs).

```
git clone https://github.com/cloudfoundry-community/cf-env apps/cf-env
cd apps/cf-env
bundle
cd -
```

Although the newly created `apps/` folder is ignored by `.gitignore` it will be correctly uploaded to Cloud Foundry as part of the deployment below.

```
godep save ./..
cf push $APPNAME --no-start
```

You now need to configure the broker with credentials for your target Cloud Foundry as an admin-level user. Most likely this will be the same Cloud Foundry you are deploying too.

```
cf set-env $APPNAME CF_API https://api.<platform_domain>
cf set-env $APPNAME CF_USER admin
cf set-env $APPNAME CF_PASS admin-password
```

Now configure how to deploy the app-as-a-service. Paths are relative to this application folder.

```
cf set-env $APPNAME CF_CATALOG_PATH ./catalog.json
cf set-env $APPNAME CF_SRC ./apps/cf-env
cf set-env $APPNAME CF_SETUP_PATH ./app/cf-env/setup.sh
cf set-env $APPNAME CF_DEP postgresql93|free,consul|free
```

CF_SETUP_PATH is script that is ran on binding service. This script is expected to return json that is added to credentials. First parameter passed into script is appname. This can be used to set environment variables for app.

Now configure how to connect to oauth app

```
cf set-env $APPNAME UI true
cf set-env $APPNAME CLIENT_ID my_client
cf set-env $APPNAME CLIENT_SECRET my_secret
cf set-env $APPNAME REDIRECT_URL https://my-client.<platform_domain>
cf set-env $APPNAME AUTH_URL https://login.<platform_domain>/oauth/authorize
cf set-env $APPNAME TOKEN_URL https://uaa.<platform_domain>/oauth/token
cf set-env $APPNAME API_URL https://api.<platform_domain>/

```

Optional configuration:

-	Skip SSL validation with CF API: `cf set-env $APPNAME CF_API_SKIP_SSL_VALID true`
-	Enable debugging: `cf set-env $APPNAME CF_DEBUG true`

To start or restart the application after any configuration changes:

```
cf restart $APPNAME
```

To register the broker:

```
export SERVICE_URL=$(cf app $APPNAME | grep urls: | awk '{print $2}')
cf create-service-broker $SERVICE admin admin https://$SERVICE_URL
cf enable-service-access cf-env
```

The latter command will make the service available to all organizations. You might want to restrict it to a subset of organizations with the `-o` flag.


Tips for apps
-----------------------

### Setup script tips

#### Parsing services attached environment variables

Appname is passed in as first environment variable. We can use it to find app GUID with that we are able to get environment variables. Using [jq](http://stedolan.github.io/jq/)

```
app_guid=$($cf curl "/v2/apps?q=name:$appname" | $jq -r '.resources | .[0].metadata.guid')
postgres_uri=$($cf curl /v2/apps/$app_guid/env | $jq -r '.system_env_json.VCAP_SERVICES.postgresql93 | .[0].credentials.uri')
```

#### Running single postgres command

[Pgopher](https://github.com/longnguyen11288/pgopher) is a static binary used to run single queries against postgres

example
```
pgopher --uri $postgres_uri -q "insert into users values(1, '$username', '$password', now(), now())"
```


### Golang tips

Using golang apps requires you to pull in app dependency of the app that is being launched.

```
godep save ./...
```

The app that is being launched need a new godep save in order to be launched as well.
