var app = angular.module('appLauncher', ['ui.router']);

app.run(['$rootScope', '$state', '$stateParams', function ($rootScope,$state,$stateParams) {
  $rootScope.$state = $state;
  $rootScope.$stateParams = $stateParams;
}])

app.config(function($stateProvider, $urlRouterProvider) {
  //
  // For any unmatched url, redirect to /state1
  $urlRouterProvider.otherwise("/home");
  //
  // Now set up the states
  $stateProvider
    .state('home', {
      url: "/home",
      templateUrl: "partials/home.html",
      controller: 'HomeCtrl'
    })
    .state('app-new', {
      url: "/app/new",
      templateUrl: "partials/new.html",
      controller: 'AppNewCtrl'
    })
    .state('app-provision', {
      url: "/app/provision/:id",
      templateUrl: "partials/provision.html",
      controller: 'AppProvisionCtrl'
    })
    .state('app-view', {
      url: "/app/view/:id",
      templateUrl: "partials/view.html",
      controller: 'AppViewCtrl'
    });
});
