String.prototype.toTitleCase = function(){
  return this.replace(/\w\S*/g, function(str){
    return str.charAt(0).toUpperCase() + str.substr(1).toLowerCase();
  });
};

String.prototype.titleize = function(){
  var overrides = {
    'noc':      'NOC',
    'centos':   'CentOS',
    'redhat':   'RedHat',
    'pam':      'PAM',
    'ldap':     'LDAP',
    'grub4dos': 'GRUB4DOS'
  };

  if(overrides.hasOwnProperty(this.toLowerCase()))
    return overrides[this.toLowerCase()];

  return this.replace(/_/g, ' ').toTitleCase();
};

var arrayDiff = function(array, a) {
  return array.filter(function(i) {return !(a.indexOf(i) > -1);});
};


var propertyGet = function(obj, path, defval){
  if(!angular.isArray(path)){
    path = path.split('.');
  }

  var work = obj;

  for(var i = 0; i < path.length; i++){
    if(work.hasOwnProperty(path[i])){
      work = work[path[i]];
    }else{
      return defval;
    }
  }


  if(angular.isDefined(work) && work != null){
    return work;
  }else{
    return defval;
  }
}


arrayMax = function(array) {
  return Math.max.apply(null, array);
};

arrayMin = function(array) {
  return Math.min.apply(null, array);
};

angular.module('coreFilters', ['ng']).
filter('titleize', function(){
  return function(text){
    if(text) return text.toString().titleize();
    return text;
  };
}).
filter('autosize', function(){
  return function(bytes,fixTo,fuzz){
    bytes = parseInt(bytes);
    fuzz = (angular.isUndefined(fuzz) ? 0.99 : +fuzz);
    if(angular.isUndefined(fuzz)) fixTo = 2;

    if(bytes >=   (Math.pow(1024,8) * fuzz))
      return (bytes / Math.pow(1024,8)).toFixed(fixTo) + ' YiB';

    else if(bytes >=   (Math.pow(1024,7) * fuzz))
      return (bytes / Math.pow(1024,7)).toFixed(fixTo) + ' ZiB';

    else if(bytes >=   (Math.pow(1024,6) * fuzz))
      return (bytes / Math.pow(1024,6)).toFixed(fixTo) + ' EiB';

    else if(bytes >=   (Math.pow(1024,5) * fuzz))
      return (bytes / Math.pow(1024,5)).toFixed(fixTo) + ' PiB';

    else if(bytes >=   (Math.pow(1024,4) * fuzz))
      return (bytes / Math.pow(1024,4)).toFixed(fixTo) + ' TiB';

    else if(bytes >=   (1073741824 * fuzz))
      return (bytes / 1073741824).toFixed(fixTo) + ' GiB';

    else if(bytes >=   (1048576 * fuzz))
      return (bytes / 1048576).toFixed(fixTo) + ' MiB';

    else if(bytes >=   (1024 * fuzz))
      return (bytes / 1024).toFixed(fixTo) + ' KiB';

    else
      return bytes + ' bytes';
  }
}).
filter('autospeed', function(){
  return function(speed,unit,fixTo,fuzz){
    speed = parseInt(speed);
    fuzz = (angular.isUndefined(fuzz) ? 0.99 : +fuzz);
    if(angular.isUndefined(fuzz)) fixTo = 2;

    if(unit){
      switch(unit.toUpperCase()){
      case 'K':
        speed = speed * 1000;
        break;
      case 'M':
        speed = speed * 1000000;
        break;
      case 'G':
        speed = speed * 1000000000;
        break;
      case 'T':
        speed = speed * 1000000000000;
        break;
      }
    }

    if(speed >= 1000000000000*fuzz)
      return (speed/1000000000000).toFixed(fixTo)+' THz';

    else if(speed >= 1000000000*fuzz)
      return (speed/1000000000).toFixed(fixTo)+' GHz';

    else if(speed >= 1000000*fuzz)
      return (speed/1000000).toFixed(fixTo)+' MHz';

    else if(speed >= 1000*fuzz)
      return (speed/1000).toFixed(fixTo)+' KHz';

    else
      return speed.toFixed(fixTo) + ' Hz';
  };
}).
filter('fix', function(){
  return function(number, fixTo){
    return parseFloat(number).toFixed(parseInt(fixTo));
  }
}).
filter('timeAgo', function(){
  return function(date){
    return moment(Date.parse(date)).fromNow();
  };
}).
filter('timeFormat', function(){
  return function(date,format,inputUnit){
    if(angular.isUndefined(date) || date == null){
      date = new Date();
    }

    if(angular.isNumber(date)){
      return moment(date).startOf('day').add((inputUnit || 'seconds'), date).format(format);
    }else{
      return moment(date).format(format);
    }
  };
}).
filter('timeAgoHuman', function(){
  return function(time,format,start){
    if(angular.isUndefined(format)){
      format = '[%Y years, ][%M months, ][%D days, ]%02h:%02m:%02s';
    }

    if(angular.isDefined(start)){
      var start = moment(start);
    }else{
      var start = moment();
    }

    var millisecondsAgo = start.diff(time);
    var units = {
      years:   (1000 * 60 * 60 * 24 * 365),
      months:  (1000 * 60 * 60 * 24 * 30),
      days:    (1000 * 60 * 60 * 24),
      hours:   (1000 * 60 * 60),
      minutes: (1000 * 60),
      seconds: 1000
    };

    var processingOrder = [
      ['Y',  'years'],
      ['M',  'months'],
      ['D',  'days'],
      ['h',  'hours'],
      ['m',  'minutes'],
      ['s',  'seconds'],
      ['ms', 'msec']
    ];

    var rv = [];


//  loop through the processing order of supported units of time
//  for each unit, find and replace the relevant part of the format
//  string (if any) with the formatted unit of time, then reduce the
//  current time difference by that unit so the next iteration does not
//  include it in the calculation
//
//  e.g.:  3,650,000 milliseconds
//
//    years?   0
//    months?  0
//    days?    0
//    hours?   :=  INT(3,650,000 / 3,600,000) (# milliseconds in an hour) == 1
//                 set current count to (3,650,000 - (3,600,000 * 1))     -> 50,000
//    minutes? 0 (50,000 is less than 1000*60 [60,000])
//    seconds? :=  INT(50,000 / 1000) (# of milliseconds is one seconds)  == 50
//
//    RESULT:  1 hour, 50 seconds
//
    for(var i = 0; i < processingOrder.length; i++){
      var token = processingOrder[i][0];
      var unit  = units[processingOrder[i][1]];
      var rx    = new RegExp('\\[?%(.[0-9]+)?'+token+'(?:(.*?)\\])?');
      var match = format.match(rx);


      if(match && millisecondsAgo >= unit){
        var number = parseInt(millisecondsAgo / unit);

        if(angular.isDefined(number)){
      //  padding to n places with given character
          if(angular.isString(match[1]) && match[1].length >= 2){
            var spaces = parseInt(match[1].slice(1));
            var padchar = match[1].slice(0,1);
            number = String(Array(spaces).join(padchar)+number.toString()).slice(-1*spaces);
          }

      //  replace matched part of format string with the formatted version
          format = format.replace(rx, number.toString()+(match[2] || ''));
        }

        millisecondsAgo = millisecondsAgo - (parseInt(millisecondsAgo / unit)*unit);
      }

    }

    return format.replace(/\[.*?\]/g,'');
  }
}).
filter('section', function(){
  return function(str, delim, start, len){
    if(str){
      var rv = str.split(delim);
      start = parseInt(start);
      len = parseInt(len);

      if($.isNumeric(start)){
        if($.isNumeric(len)){
          return rv.slice(start, len).join(delim);
        }

        return rv.slice(start).join(delim);
      }

      return str;
    }

    return null;
  };
}).
filter('jsonify', function () {
  return function(obj, indent){
    return JSON.stringify(obj, null, (indent || 4));
  };
}).
config(['$provide', function($provide) {
  $provide.factory('truncateFilter', function(){
    return function(text, length, start){
      if(angular.isUndefined(start)){
        start = 0;
      }

      if(text.length <= length || start > test.length){
        return text;
      }else{
        return String(text).substring(start, length);
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('skipFilter', function(){
    return function(array, skip){
      if (!(array instanceof Array)) return array;
      skip = parseInt(skip);
      rv = [];

      if(skip > array.length) return [];

      for(var i = skip; i < array.length; i++){
        rv.push(array[i]);
      }

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('sliceFilter', function(){
    return function(array, start, end){
      if (!(array instanceof Array)) return array;
      return array.slice((start || 0), end);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('sumFilter', function(){
    return function(array){
      if (!(array instanceof Array)) return NaN;
      rv = 0.0;

      for(var i = 0; i < array.length; i++){
        rv += parseFloat(array[i]);
      }

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('joinFilter', function(){
    return function(array, delimiter){
      if (!(array instanceof Array)) return array;
      if(!delimiter) delimiter = '';
      return array.join(delimiter);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('emptyFilter', function(){
    return function(array,key){
      if (!(array instanceof Array)) return array;
      rv = array.filter(function(i){
        if($.isPlainObject(i))
          return i.hasOwnProperty(key) && !i[key];
        else if(typeof(i) == 'string')
          return (i.length != 0);
        else
          return !i;
      });
      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('compactFilter', function(){
    return function(array,key){
      if (!(array instanceof Array)) return array;

      rv = array.filter(function(i){
        if(angular.isObject(i))
          return i.hasOwnProperty(key) && i[key];
        else if(typeof(i) == 'string')
          return (i.length == 0);
        else if(angular.isUndefined(i) || i == null)
          return false;
        else
          return i;
      });

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('propertyFilter', function(){
    return function(array,key,value,exclude){
      if (!(array instanceof Array)) return array;
      rv = array.filter(function(i){
        if(angular.isObject(i)){
          var v = (exclude ? !i.hasOwnProperty(key) : i.hasOwnProperty(key));

          if(v && typeof(value) != 'null' && typeof(value) != 'undefined' && i[key] == value){
            return true;
          }else{
            return false;
          }
        }else{
          return array;
        }
      });
      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('flattenFilter', function(){
    return function(array){
      if (!(array instanceof Array)) return array;

      return array.reduce(function(a,b){
        return a.concat(b);
      });
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('pluckFilter', function(){
    return function(array,key,defval){
      if (!(array instanceof Array)) return array;

      rv = []

      for(var i = 0; i < array.length; i++){
        if(angular.isObject(array[i])){
          rv.push(propertyGet(array[i], key, defval));
        }
      }

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('containsFilter', function(){
    return function(array,value,strict){
      if (!(array instanceof Array))
        return false;

  //  strict type checking (default:true)
      if(strict === false){
        for(var i = 0; i < array.length; i++){
          if(array[i].toString() == value.toString()){
            return true;
          }
        }

        return false;
      }else{
        return (array.indexOf(value) > -1);
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('diffFilter', function(){
    return function(array,other){
      if (!(array instanceof Array)) array = [];
      if (!(other instanceof Array)) other = [];

      return arrayDiff(array, other);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('lengthFilter', function(){
    return function(obj){
      if(obj instanceof Array){
        return obj.length;
      }else if($.isPlainObject(obj)){
        return Object.keys(obj).length;
      }else if(typeof(obj) == 'string'){
        return obj.length;
      }else{
        return null;
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('minFilter', function(){
    return function(obj,field){
      if(obj instanceof Array){
        return arrayMin(obj);
      }else if($.isPlainObject(obj) && angular.isDefined(field) && obj[field] instanceof Array){
        return arrayMin(obj[field]);
      }else{
        return null;
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('maxFilter', function(){
    return function(obj,field){
      if(obj instanceof Array){
        return arrayMax(obj);
      }else if($.isPlainObject(obj) && angular.isDefined(field) && obj[field] instanceof Array){
        return arrayMax(obj[field]);
      }else{
        return null;
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('fillArrayFilter', function(){
    return function(upper,lower){
      var rv = [];

      if(angular.isUndefined(lower)){
        lower = 0;
      }else{
        lower = parseInt(lower);
      }

      upper = parseInt(upper);

      for(var i = lower; i < (upper+lower); i++){
        rv.push(i);
      }

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('rangeFilter', function(){
    return function(input, total, offset) {
      total = parseFloat(total);

      if(angular.isUndefined(offset)){
        offset = 0;
      }

      for (var i=offset; i<(total+offset); i++){
        input.push(i);
      }

      return input;
    };
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('startsWithFilter', function(){
    return function(obj,test,ci){
      if(angular.isUndefined(obj) || obj === null){
        return false;
      }

      return obj.toString().match(new RegExp("^"+test,(ci==true ? "g" : undefined)));
    };
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('endsWithFilter', function(){
    return function(obj,test,ci){
      if(angular.isUndefined(obj) || obj === null){
        return false;
      }

      return obj.toString().match(new RegExp(test+"$",(ci==true ? "g" : undefined)));
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isDefinedFilter', function(){
    return function(obj){
      return angular.isDefined(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isUndefinedFilter', function(){
    return function(obj){
      return angular.isUndefined(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isObjectFilter', function(){
    return function(obj){
      return angular.isObject(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isArrayFilter', function(){
    return function(obj){
      return angular.isArray(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isStringFilter', function(){
    return function(obj){
      return angular.isString(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isDateFilter', function(){
    return function(obj){
      return angular.isDate(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isNumberFilter', function(){
    return function(obj){
      return angular.isNumber(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isFunctionFilter', function(){
    return function(obj){
      return angular.isFunction(obj);
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('isEmptyFilter', function(){
    return function(obj, trace){
      if(angular.isString(obj) && obj.length == 0){
        return true;
      }else if(angular.isObject(obj) && obj.length == 0){
        return true;
      }else if(angular.isUndefined(obj)){
        return true;
      }else if(obj === null){
        return true;
      }

      return false;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('replaceFilter', function(){
    return function(str,find,rep,all){
      if(str instanceof Array){
        for(var i in str){
          if(typeof(str[i]) == 'string'){
            if(all == true){
              str[i] = str[i].replace(new RegExp(find,'g'), rep);
            }else{
              str[i] = str[i].replace(find, rep);
            }
          }
        }

        return str;
      }else if(typeof(str) == 'string'){
        if(all == true){
          return str.replace(new RegExp(find,'g'), rep);
        }

        return str.replace(find, rep);
      }else{
        return str;
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('selectFilter', function(){
    return function(obj,value){
      if(angular.isArray(obj)){
        return obj.filter(function(i){
          return i==value;
        });
      }else if(angular.isObject(obj)){
        var rv = {};
        angular.forEach(obj, function(v,k){
          if(v==value){
            rv[k] = v;
          }
        });

        return rv;
      }else{
        return obj;
      }
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('keysFilter', function(){
    return function(obj){
      var rv = [];

      if(angular.isArray(obj) && angular.isObject(obj[0])){
        angular.forEach(obj, function(value, idx){
          rv = rv.concat(Object.keys(value));
        });
      }else if(angular.isObject(obj)){
        rv = Object.keys(obj);
      }

      return rv;
    }
  });
}]).
config(['$provide', function($provide) {
  $provide.factory('valuesFilter', function(){
    return function(obj){
      var rv = [];
      if(angular.isObject(obj)){
        angular.forEach(obj, function(v){
          rv.push(v);
        });
      }

      return rv;
    }
  });
}]);