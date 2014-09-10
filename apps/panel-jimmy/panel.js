angular.module('app', [
  'ng'
])
.run(['$rootScope', '$http', function($rootScope, $http){
  $rootScope.api = {
    basehost: 'localhost:9521',

    head: function(endpoint, config){
      return $http.head('http://'+$rootScope.api.basehost+endpoint, config);
    },

    get: function(endpoint, config){
      return $http.get('http://'+$rootScope.api.basehost+endpoint, config);
    },

    put: function(endpoint, data){
      return $http.put('http://'+$rootScope.api.basehost+endpoint, data);
    },

    post: function(endpoint, data){
      return $http.post('http://'+$rootScope.api.basehost+endpoint, data);
    },

    delete: function(endpoint, config){
      return $http.delete('http://'+$rootScope.api.basehost+endpoint, config);
    }
  };

  $rootScope.bus = {
    connect: function(){
      $rootScope.bus._connection = new WebSocket('ws://'+$rootScope.api.basehost+'/v1/bus');

  //  wire up handlers (if defined)
      angular.forEach(['open', 'close', 'message'], function(verb){
        if(angular.isFunction($rootScope.bus['on'+verb])){
          $rootScope.bus._connection['on'+verb] = function(e){
            $rootScope.$apply(function(){
              $rootScope.bus['on'+verb](angular.fromJson(e.data), e);
            });
          }
        }
      });
    }
  };
}])
.controller('PanelController', ['$scope', '$interval', '$rootScope', function($scope, $interval, $rootScope){
  $rootScope.bus.onmessage = function(data, raw_event){
    console.debug("msg", data, raw_event);
    $scope.reload();
  }

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

  $rootScope.bus.connect();
  $scope.reload();
}]);

function TimeCtrl($scope, $timeout) {
    $scope.clock = "10:00:00 AM"; // initialise the time variable
    $scope.tickInterval = 1000 //ms

    var tick = function() {
        $scope.clock = Date.now() // get the current time
        $timeout(tick, $scope.tickInterval); // reset the timer
    }

    // Start the timer
    $timeout(tick, $scope.tickInterval);
};

$( "#search" ).hover(
  function() {
    $( this ).animate({width: '160px'}, 100);
  }, function() {
    $( this ).animate({width: '30px'}, 100);
  }
);
