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
