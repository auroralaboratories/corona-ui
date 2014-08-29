angular.module('app', [
  'ng'
])
.run(['$rootScope', '$http', function($rootScope, $http){
  $rootScope.api = {
    baseuri: 'http://localhost:9521',

    head: function(endpoint, config){
      return $http.head($rootScope.api.baseuri+endpoint, config);
    },

    get: function(endpoint, config){
      return $http.get($rootScope.api.baseuri+endpoint, config);
    },

    put: function(endpoint, data){
      return $http.put($rootScope.api.baseuri+endpoint, data);
    },

    post: function(endpoint, data){
      return $http.post($rootScope.api.baseuri+endpoint, data);
    },

    delete: function(endpoint, config){
      return $http.delete($rootScope.api.baseuri+endpoint, config);
    }
  };
}])
.controller('PanelController', ['$scope', '$interval', function($scope, $interval){
  $scope.reload = function(){
    $scope.api.get('/v1/session/workspaces').success(function(data){
      $scope.workspaces = data;
    });

    $scope.api.get('/v1/session/windows').success(function(data){
      $scope.windows = data;
    });
  }

  $scope.actionWindow = function(id, action){
    $scope.api.put('/v1/session/windows/'+id+'/'+action).success(function(){
      $scope.reload();
    })
  }

  $scope.reload();
}]);
