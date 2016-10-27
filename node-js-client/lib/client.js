/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

var Promise = require('bluebird');
var Joi = require('joi');
var request = Promise.promisify(require('request'));

var clientConfigSchema = Joi.object().keys({
  host: Joi.string().hostname().required(),
  port: Joi.number().integer().min(1).max(1 << 16).required(),
  id: Joi.string().required(),
  key: Joi.string().required(),
  timeout: Joi.number().integer().positive()
});

/**
 * @class IPReputationServiceClient
 * @constructor
 * @param {Object} config
 *   @param {String} config.host
 *   @param {Number} config.port
 *   @param {String} config.id id for the HAWK header
 *   @param {String} config.key key for the HAWK header
 *   @param {Number} config.timeout positive integer of the number of milliseconds to wait for a server to send response headers (passed as parameter of the same name to https://github.com/request/request)
 * @return {IPReputationServiceClient}
 */
var client = function(config) {
  Joi.assert(config, clientConfigSchema);

  this.baseUrl = 'http://' + config.host + ':' + config.port + '/';

  this.credentials = {
    id: config.id,
    key: config.key,
    algorithm: 'sha256'
  };

  this.timeout = config.timeout || 30 * 1000;
  return this;
};

/**
 * @method get
 * @param {String} an IP address to fetch reputation for
 * @return {Promise}
 */
client.prototype.get = function (ip) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
    hawk: {
      credentials: this.credentials
    },
    timeout: this.timeout
  };

  return request(requestOptions);
};


/**
 * @method add
 * @param {String} an IP address to assign a reputation
 * @param {Number} a reputation/trust value from 0 to 100 inclusive (higher is more trustworthy)
 * @return {Promise}
 */
client.prototype.add = function (ip, reputation) {
  var requestOptions = {
    uri: this.baseUrl,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    hawk: {
      credentials: this.credentials
    },
    body: JSON.stringify({'ip': ip, 'reputation': reputation}),
    timeout: this.timeout
  };

  return request(requestOptions);
};

/**
 * @method update
 * @param {String} an IP address to change a reputation for
 * @param {Number} a reputation/trust value from 0 to 100 inclusive (higher is more trustworthy)
 * @return {Promise}
 */
client.prototype.update = function (ip, reputation) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    hawk: {
      credentials: this.credentials
    },
    body: JSON.stringify({'reputation': reputation}),
    timeout: this.timeout
  };

  return request(requestOptions);
};

/**
 * @method remove
 * @param {String} an IP address to remove an associated reputation for
 * @return {Promise}
 */
client.prototype.remove = function (ip) {
  var requestOptions = {
    uri: this.baseUrl + ip,
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
    hawk: {
      credentials: this.credentials
    },
    timeout: this.timeout
  };

  return request(requestOptions);
};

module.exports = client;
