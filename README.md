App Launching Service Broker for Cloud Foundry
==============================================

A service broker to provision an application, including dependent service instances.

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
cd back/to/app-launching-service-broker
source bin/env.sh
```

This will create a set of environment variables used to configure the service broker:

```
$ env | grep CF
CF_API=https://api.gotapaas.com
CF_SRC=/Users/drnic/Projects/cloudfoundry/apps/cf-env
CF_DEP=postgresql93|free,consul|free
CF_CAT=./catalog.json
```

You now need to configure which admin user credentials the broker will use to communicate with Cloud Foundry:

```
export CF_USER=admin
export CF_PASS=c1oudc0w
```

If you are using self-signed certificates, you may need to ignore SSL verification:

```
export CF_API_SKIP_SSL_VALID=true
```

You can now run the service broker locally:

```
go run main.go
```

The broker will be running on port 9999 by default.

```
$ curl http://localhost:9999/v2/catalog
{"services":[{"id":"B6D73C9E-302D-4B78-BC46-56E92C6C000D","name":"cf-env","description":"View environment variables","bindable":true,"tags":["demo","backend"],"plans":[{"id":"4672FA24-4330-404B-AFC0-235AB6EA0F8C","name":"simple","description":"Simple","free":true}]}]}
```

The output here matches the contents of the `./catalog.json` example file.

You can now even register your local app with a remote Cloud Foundry using [ngrok](https://ngrok.com/). Run the following in another terminal:

```
$ ngrok 9999
```

It will display a public URL for your local broker app.

Register your broker app URL:

```
$ cf create-service-broker cf-env-broker admin admin http://3f1c1555.ngrok.com
$ cf enable-service-access cf-env-broker
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
