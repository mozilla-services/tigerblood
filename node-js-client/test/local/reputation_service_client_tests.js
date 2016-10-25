/* Any copyright is dedicated to the Public Domain.
 * http://creativecommons.org/publicdomain/zero/1.0/ */

var test = require('tap').test
var IPReputationClient = require('../../lib/client')
var client = new IPReputationClient({
  host: '127.0.0.1',
  port: 8080,
  id: 'root',
  key: 'toor'
})

test(
  'throws exception when missing required config values',
  function (t) {
    t.throws(function () { new IPReputationClient({}) })
    t.throws(function () { new IPReputationClient({port: 8080, id: 'root', key: 'toor'}) })
    t.throws(function () { new IPReputationClient({host: '127.0.0.1', id: 'root', key: 'toor'}) })
    t.throws(function () { new IPReputationClient({host: '127.0.0.1', port: 8080, key: 'toor'}) })
    t.throws(function () { new IPReputationClient({host: '127.0.0.1', port: 8080, id: 'root'}) })
    t.end()
  }
)

test(
  'does not get reputation for a nonexistent IP',
  function (t) {
    client.get('127.0.0.1', function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 404)
      t.end()
    })
  }
)

test(
  'does not update reputation for nonexistent IP',
  function (t) {
    client.update('127.0.0.1', 5, function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 200)
      t.equal(body, '')
      t.end()
    })
  }
)

test(
  'does not remove reputation for a nonexistent IP',
  function (t) {
    client.remove('127.0.0.1', function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 200)
      t.equal(body, '')
      t.end()
    })
  }
)


// the following tests need to run in order

test(
  'adds reputation for new IP',
  function (t) {
    client.add('127.0.0.1', 50, function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 201)
      t.end()
    })
  }
)

test(
  'does not add reputation for existing IP',
  function (t) {
    client.add('127.0.0.1', 50, function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 500)
      t.end()
    })
  }
)

test(
  'gets reputation for a existing IP',
  function (t) {
    client.get('127.0.0.1', function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 200)
      t.equal(body, '{"IP":"127.0.0.1","Reputation":50}')
      t.end()
    })
  }
)

test(
  'updates reputation for existing IP',
  function (t) {
    client.update('127.0.0.1', 5, function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 200)
      t.equal(body, '')
      t.end()
    })
  }
)

test(
  'removes reputation for existing IP',
  function (t) {
    client.remove('127.0.0.1', function (error, response, body) {
      t.equal(error, null)
      t.equal(response.statusCode, 200)
      t.equal(body, '')
      t.end()
    })
  }
)
