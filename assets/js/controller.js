app.controller('HomeCtrl', function($scope, $http, $state) {
})

app.controller('AppNewCtrl', function($scope, $http) {
  $http.get('/ui/organizations').success(function(data) {
    $scope.orgs = data
  })

  $scope.getSpace = function(guid){
    $http.get('/ui/organizations/' + guid + '/spaces').success(function(data) {
      $scope.spaces = data
    })
  }

  $scope.setProjectSpaceName = function(spaceName){
    $scope.projectSpaceName = spaceName
  }
})

app.controller('AppViewCtrl', function($scope, $http) {
})

app.controller('AppProvisionStatusCtrl', function($scope, $http) {
})
