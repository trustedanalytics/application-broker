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

app.controller('AppProvisionCtrl', function($scope, $http, $stateParams, $timeout) {
  id = $stateParams.id
  var levels = {
    creating: 1,
    uploading: 2,
    provision_cloudera: 3,
    provision_db: 4,
    bind_services: 5,
    restarting_atk: 6,
    create_user: 7
  }

  $scope.findLevel = function(level) {
    if (level == $scope.level) {
      return "list-group-item-info"
    } else if (level < $scope.level) {
      return "list-group-item-success"
    } else {
      return ""
    }
  }

  //5 second timer to refresh app and set level
  var getApp = function() {
    $http.get('/ui/apps/' + id ).success(function(data) {
      $scope.app = data
      $scope.level = levels[$scope.app.environment_json["APP_LAUNCHER_STATE"]]
    })
    $timeout(getApp, 5000);
  }

  var promise = $timeout(getApp, 5000);
  getApp();

})

app.controller('NavBarCtrl', function($scope, $http) {
  $http.get('/ui/apps').success(function(data) {
    $scope.apps = data
  })
})
