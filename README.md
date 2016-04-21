[![Build Status](https://travis-ci.org/trustedanalytics/application-broker.svg)](https://travis-ci.org/trustedanalytics/application-broker)

Application Broker for Cloud Foundry
==============================================

A service broker to provision an application stack - set of applications, services and user provided services.

**This is a third version, which support copying full application stacks, binding same services to multiple apps in single stack and generating random passwords when copying source ups with e.g. $RANDOM16.**
**Second version used REST API of CF to retrieve and add data.**
**First implementation was based on cf/cli calls which was considered bad solution due to concurrency problems. Now we are using regular REST requests under the hood.**


Idea behind
-----------
Cloud Foundry introduces two notions when talking about software running within it. **Application** (pushed and controlled by developers) and **Service** (spawned by so-called brokers and used by Applications). In general, the first one is a kind of web service. The second one, works as a resource (database for example) to be used on-demand by Application. Initially Cloud Foundry provides a set of Service offerings to choose from (postgresql, mongodb, NATS, etc...). Every single offering in internal marketplace has broker behind it. Broker manages every instance of service and handles its binding to Applications. No service exists without its corresponding broker.

To easily create service offerings without implementing separate broker you may want to use **Application Broker**. The only thing you need to do is to prepare reference app and register it in our broker. Then you will be able to spawn copies of your reference app and treat them like service instances. Offering will be also visible within CF marketplace. As simple as that.

References:
[Custom Services](https://docs.cloudfoundry.org/services/)

Features 
--------

* Publishing application stack in marketplace

[![Example stack](https://github.com/intel-data/application-broker/blob/master/app_in_marketplace.png)](https://github.com/intel-data/application-broker/blob/master/app_in_marketplace.png)

* Cloning existing stacks of apps, services and user provided services if they create a DAG with single root

[![Example stack](https://github.com/intel-data/app-dependency-discoverer/blob/master/example_tree.png)](https://github.com/intel-data/app-dependency-discoverer/blob/master/example_tree.png)

* Linking apps by adding user provided service as JSON containing url field with value http://<host>.<domain>. App mentioned has to be in the same space as source app. Cloned user provided service contains updated url pointing cloned application. Convention for user provided services names linking apps is <host>-ups.

```
cf cups app1-ups -p url

url> http://app1.<domain>
```

* Replacing $RANDOM8, $RANDOM16, $RANDOM24 and $RANDOM32 in user provided services copied with random string of width 8, 16, 24 or 32 from characterset [A-Za-z0-9]. This can be used to generate new password for every stack spawned. Parts for replacement can be on different levels of JSON in UPS. Replacement is done when copying so original user provided service is using phrase with $.

```
{
    "credentials": {
     "password": "$RANDOM16" -> "hTn8X07zm8KRDKr1"
    },
    "label": "user-provided",
    "name": "random-ups",
    "syslog_drain_url": "",
    "tags": []
}
```


Usage
-----

Application broker is regular long-living app that has to be pushed to CF. An example of properly filled manifest.yml shall look like this:
```
---
applications:
- name: application-broker
  memory: 256M
  instances: 1
  path: .
  buildpack: go_buildpack
  services:
    - application-broker-mongodb
  env:
    AUTH_USER: admin            #broker API is secured with these BasicAuth credentials
    AUTH_PASS: password
    CLIENT_ID: client           #broker communicates with CF using OAuth
    CLIENT_SECRET: clientSecret #when asking for token it uses these credentials
    TOKEN_URL: https://uaa.yourserver.com/token_key
    CF_API: http://api.yourserver.com
    VERSION: "0.5.8"
```

Application broker is using app dependency discoverer for application stack discovery. It needs url of that app in url field of app-dependency-discoverer-ups user provided service, as well as auth_user and auth_pass for basic authentication. For local run http://localhost:9998 is taken.
Create user provided service with command:
```
cf cups app-dependency-discoverer-ups -p "{\"auth_pass\": \"<password>\", \"auth_user\": \"<user>\", \"url\": \"http://<hostname>.<domain>\" }"
```

When manifest.yml is ready and ups with required name is available in selected space, the following command can be issued:
```
$ cf push
```

Application broker serves a catalog of service offerings. It means that it is responsible for multiple entries in marketplace. Informations about services that broker controls (catalog) are held in mongodb database bound to ApplicationBroker. Initially catalog is empty. You must register at least one to benefit from using it. So, push an app that you want to place in marketplace and remember its guid (we will call it referenceAppGuid).

Next, you should register your app within Application Broker catalog using Catalog API.
For example (as simple as possible):
```
curl -sL $APPLICATION_BROKER_ADDRESS/v2/catalog -X POST  \
    -u $AUTH_USER:$AUTH_PASS                             \
    -H "Content-Type: application/json"                  \
    -d '{
            "app" : {"metadata" : {"guid" : "<referenceAppGuid>"}},
            "id" : "<place random guid here>",
            "plans" : [{"id" : "<place random guid here>"}],
            "description" : "<describe your service briefly>",
            "name" : "<service exposed by your broker>"
        }'
```
Now Application Broker has one service registered. When asked it responds with non-empty catalog. You can check by firing:
```
curl -sL $APPLICATION_BROKER_ADDRESS/v2/catalog -X GET -u $AUTH_USER:$AUTH_PASS
```
Next we need to inform CF that your Application Broker instance is in fact broker.
```
$ cf create-service-broker <brokerName> admin admin http://address.of.pushed.application.broker
$ cf enable-service-access <service exposed by your broker>
```
While running `cf create-service-broker` Cloud Controller make request to provided URL and saves catalog that Application Broker exposes.

Once the service broker is running as an app, registered as a service broker, and enabled for access by users, a user can use like any other service broker:
```
cf create-service <service exposed by your broker> Simple <instanceName>
cf bind-service my-app <instanceName>
cf restage my-app
```

The client application `my-app` would have a `$VCAP_SERVICE` service for `<instanceName>`. The credentials would include the url to access the backend service application.

Behind the scenes, when `cf create-service` was invoked the CLI asked the Cloud Controller for the new service instance. The Cloud Controller asked the Application Broker for a new service instance. The Application Broker deploys a new backend application as a copy of referenceApp.

If the backend application requires any services for itself, then those too are created and bound to the new backend application.

The backend application and its own dependency services are created into the same organization and space being used by the end user.

NATS
-----------
Application Broker uses NATS messagebus to emit events. For now, events are being sent on every service instance provisioning. Events are meant to inform users about correct or erroneous results of operation. To enable NATS for your broker use the environment variable named `NATS_URL` pointing to address your NATS is listening on. Additionally, you can specify topic Application Broker should talk on using `NATS_SERVICE_CREATION_SUBJECT`. By default it is `service-creation`.

Development
-----------

### Prerequisites

Application broker is using [app-dependency-discoverer] (https://github.com/intel-data/app-dependency-discoverer) so you need to run it locally before using application-broker. You also need to specify url and credentials to it in app-dependency-discoverer-ups with content described earlier.

To locally develop this service broker, we encourage you to use lightweight reference stack that will push and start fast. Testing won't take too much time.

Additionally you will need mongodb instance. Install it by using package-manager your distro provides. For Ubuntu/Debian it will be: `sudo apt-get install mongodb`. Local Application Broker will connect to it on default port so no additional configuration is needed.

### Running locally

Install project dependencies:
```
godep restore ./...
```

You can run the Application Broker locally via [gin](https://github.com/codegangsta/gin), which will automatically reload any file changes during development. Our application needs several environment variables to work properly. Ensure that they are exported before starting. Running following commands is sufficient to run the Application Broker correctly:

```
go get github.com/codegangsta/gin
export CF_API=http://api.yourserver.com
export TOKEN_URL=https://uaa.yourserver.com/oauth/token
export CLIENT_ID=client
export CLIENT_SECRET=clientSecret
export AUTH_USER=admin
export AUTH_PASS=password
gin -a 9999 -i run main.go
```

The broker, via `gin`, will be running on port 3000 by default.

### Unit Testing

To write self-describing unit tests easily we adopted ginkgo framework. Running these is simple as:

```
go get github.com/onsi/ginkgo/ginkgo
export PATH=$PATH:$GOPATH/bin
#commands above need to be executed just once
ginkgo -r
```

### IDE
We recommend using [IntelliJ IDEA](https://www.jetbrains.com/idea/) as IDE with [golang plugin](https://github.com/go-lang-plugin-org/go-lang-idea-plugin). To apply formatting automatically on every save you may use go-fmt with [File Watcher plugin](http://www.idmworks.com/blog/entry/automatically-calling-go-fmt-from-intellij).


Tips
-----------------------

### Golang tips

Developing golang apps requires you store all dependencies (Godeps) in separate directory. They shall be placed in source control.

```
godep save ./...
```

Command above places all dependencies from `$GOPATH`, your app uses, in Godeps and writes its versions to Godeps/Godeps.json file.


Limitations
-----------------------
Additionally, in special circumstances, some problems may occur when spawning new instance with dependencies. Imagine referenceApp with dependencyServiceInstance bound to it. It is possible to spawn copy of referenceApp to space that dependencyService is not enabled in. In such situation provision operation will end up with failure.
