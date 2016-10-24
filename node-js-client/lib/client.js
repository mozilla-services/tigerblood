/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

var request = require('request')
var hawk = require('hawk')

var generateHawkHeader = function (credentials, requestOptions) {
  var header = hawk.client.header(requestOptions.uri, requestOptions.method, {
    credentials: credentials,
    ext: '',
    contentType: 'application/json',
    payload: requestOptions.body ? requestOptions.body : ''
  })

  return header
}

/**
 * @class IPReputationClient
 * @constructor
 * @param {Object} config
 *   @param {String} config.host
 *   @param {Number} config.port
 *   @param {id} config.id id for the HAWK header
 *   @param {id} config.key key for the HAWK header
 * @return {IPReputationServiceClient}
 */
var client = function(config) {
  if (!Object.prototype.hasOwnProperty.call(config, 'host')) {
    throw new Error('Missing required param host for IP Reputation Client.')
  } else if (!Object.prototype.hasOwnProperty.call(config, 'port')) {
    throw 'Missing required param port for IP Reputation Client.'
  } else if (!Object.prototype.hasOwnProperty.call(config, 'id')) {
    throw 'Missing required param (hawk) id for IP Reputation Client.'
  } else if (!Object.prototype.hasOwnProperty.call(config, 'key')) {
    throw 'Missing required param (hawk) key for IP Reputation Client.'
  }

  this.baseUrl = 'http://' + config.host + ':' + config.port + '/'

  this.credentials = {
    id: config.id,
    key: config.key,
    algorithm: 'sha256'
  }

  return this
}

/**
 * @method get
 * @param {String} an IP address to fetch reputation for
 * @param {Function} request callback called with args error, http.IncomingMessage, and body.
 * @return {undefined}
 */
client.prototype.get = function (ip, callback) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }

  var header = generateHawkHeader(this.credentials, requestOptions)
  requestOptions.headers.Authorization = header.field

  request(requestOptions, callback)
}


/**
 * @method add
 * @param {String} an IP address to assign a reputation
 * @param {Number} a reputation/trust value from 0 to 100 inclusive (higher is more trustworthy)
 * @param {Function} request callback called with args error, http.IncomingMessage, and body.
 * @return {undefined}
 */
client.prototype.add = function (ip, reputation, callback) {
  var requestOptions = {
    uri: this.baseUrl,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({'ip': ip, 'reputation': reputation})
  }

  var header = generateHawkHeader(this.credentials, requestOptions)
  requestOptions.headers.Authorization = header.field

  request(requestOptions, callback)
}

/**
 * @method update
 * @param {String} an IP address to change a reputation for
 * @param {Number} a reputation/trust value from 0 to 100 inclusive (higher is more trustworthy)
 * @param {Function} request callback called with args error, http.IncomingMessage, and body.
 * @return {undefined}
 */
client.prototype.update = function (ip, reputation, callback) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({'reputation': reputation})
  }

  var header = generateHawkHeader(this.credentials, requestOptions)
  requestOptions.headers.Authorization = header.field

  request(requestOptions, callback)
}

/**
 * @method remove
 * @param {String} an IP address to remove an associated reputation for
 * @param {Function} request callback called with args error, http.IncomingMessage, and body.
 * @return {undefined}
 */
client.prototype.remove = function (ip, callback) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    }
  }

  var header = generateHawkHeader(this.credentials, requestOptions)
  requestOptions.headers.Authorization = header.field

  request(requestOptions, callback)
}

module.exports = client
