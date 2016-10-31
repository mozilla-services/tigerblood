### Tigerblood (IP Reputation Service) node.js client library

Client library to send IP reputations to the tigerblood service.

Usage:

Create a client:

```js
const IPReputationClient = require('ip-reputation-service-client-js')

const client = new IPReputationClient({
    ip: '<tigerblood service IP address>',
    port: '<tigerblood service port>',
    id: '<a hawk ID>',
    key: '<a hawk key>',
    timeout: <number in ms>
})

Get the reputation for an IP:

```js
client.get('127.0.0.1').then(function (response) {
    if (response && response.statusCode === 404) {
        console.log('No reputation found for 127.0.0.1');
    } else {
        console.log('127.0.0.1 has reputation: ', response.body.Reputation);
    }
});
```

Set the reputation for an IP:

```js
client.add('127.0.0.1', 20).then(function (response) {
    console.log('Added reputation of 20 for 127.0.0.1');
});
```

Update the reputation for an IP:

```js
client.update('127.0.0.1', 79).then(function (response) {
    console.log('Set reputation for 127.0.0.1 to 79.');
});
```

Remove an IP:

```js
client.remove('127.0.0.1').then(function (response) {
    console.log('Removed reputation for 127.0.0.1.');
});
```

Send a violation for an IP:

```js
client.sendViolation('127.0.0.1', 'exceeded-password-reset-failure-rate-limit').then(function (response) {
    console.log('Upserted reputation for 127.0.0.1.');
});
```

## Development

1. Create the following `config.yml` in project root:

```yml
credentials:
  root: toor
```

1. run tigerblood with a new database (`docker run --name postgres -p 127.0.0.1:5432:5432 -d postgres-ip4r`)
1. install this library with `npm install`
1. run db setup script: `scripts/setup-test-db.sh`
1. run tigerblood from the project root: `CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' ./cmd/tigerblood/ && TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" ./tigerblood`
1. run `npm test` to test the client against the tigerblood server
1. stop the tigerblood server
1. run db cleanup script: `scripts/cleanup-test-db.sh`
