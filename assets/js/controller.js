app.controller('HomeCtrl', function($scope, $http, $state) {
})

app.controller('NewCtrl', function($scope, $http) {
  $http.get('/ui/organizations').success(function(data) {
    $scope.orgs = data
  })

  $scope.getSpace = function(guid){
    $http.get('/ui/organizations/' + guid + '/spaces').success(function(data) {
      $scope.spaces = data
    })
  }
})

app.controller('ViewCtrl', function($scope, $http) {
})

app.controller('ProvisionCtrl', function($scope, $http) {
})
