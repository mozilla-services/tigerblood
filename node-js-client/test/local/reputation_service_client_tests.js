/* Any copyright is dedicated to the Public Domain.
 * http://creativecommons.org/publicdomain/zero/1.0/ */

var test = require('tap').test;
var IPReputationClient = require('../../lib/client');
var client = new IPReputationClient({
  host: '127.0.0.1',
  port: 8080,
  id: 'root',
  key: 'toor',
  timeout: 50
});

test(
  'throws exception when missing required config values',
  function (t) {
    [
      {},
      {port: 8080, id: 'root', key: 'toor'},
      {host: '127.0.0.1', id: 'root', key: 'toor'},
      {host: '127.0.0.1', port: 8080, key: 'toor'},
      {host: '127.0.0.1', port: 8080, id: 'root'}
    ].forEach(function (badConfig) {
      t.throws(function () {
        return new IPReputationClient(badConfig);
      });
    });
    t.end();
  }
);

test(
  'does not get reputation for a nonexistent IP',
  function (t) {
    client.get('127.0.0.1').then(function (response) {
      t.equal(response.statusCode, 404);
      t.end();
    });
  }
);

test(
  'does not update reputation for nonexistent IP',
  function (t) {
    client.update('127.0.0.1', 5).then(function (response) {
      t.equal(response.statusCode, 404);
      t.equal(response.body, undefined);
      t.end();
    });
  }
);

test(
  'does not remove reputation for a nonexistent IP',
  function (t) {
    client.remove('127.0.0.1').then(function (response) {
      t.equal(response.statusCode, 200);
      t.equal(response.body, undefined);
      t.end();
    });
  }
);


// the following tests need to run in order

test(
  'adds reputation for new IP',
  function (t) {
    client.add('127.0.0.1', 50).then(function (response) {
      t.equal(response.statusCode, 201);
      t.end();
    });
  }
);

test(
  'does not add reputation for existing IP',
  function (t) {
    client.add('127.0.0.1', 50).then(function (response) {
      t.equal(response.statusCode, 409);
      t.equal(response.body, 'Reputation is already set for that IP.');
      t.end();
    });
  }
);

test(
  'gets reputation for a existing IP',
  function (t) {
    client.get('127.0.0.1').then(function (response) {
      t.equal(response.statusCode, 200);
      t.deepEqual(response.body, {'IP':'127.0.0.1','Reputation':50});
      t.end();
    });
  }
);

test(
  'updates reputation for existing IP',
  function (t) {
    client.update('127.0.0.1', 5).then(function (response) {
      t.equal(response.statusCode, 200);
      t.equal(response.body, undefined);
      t.end();
    });
  }
);

test(
  'removes reputation for existing IP',
  function (t) {
    client.remove('127.0.0.1').then(function (response) {
      t.equal(response.statusCode, 200);
      t.equal(response.body, undefined);
      return client.get('127.0.0.1'); // verify removed IP is actually gone
    }).then(function (response) {
      t.equal(response.statusCode, 404);
      t.equal(response.body, undefined); // JSON.stringify() -> undefined
      t.end();
    });
  }
);

test(
  'times out a GET request',
  function (t) {
    var timeoutClient = new IPReputationClient({
      host: '10.0.0.0', // a non-routable host
      port: 8080,
      id: 'root',
      key: 'toor',
      timeout: 1 // ms
    });

    timeoutClient.get('127.0.0.1').then(function () {}, function (error) {
      t.notEqual(error.code, null);
      t.end();
    });
  }
);
