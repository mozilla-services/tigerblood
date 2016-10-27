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

## Development

1. run tigerblood with a new database (`docker run --name postgres -p 127.0.0.1:5432:5432 -d postgres-ip4r`) and the following `config.yml`:

```yml
credentials:
  root: toor
```

2. install this library with `npm install`
3. run `npm test` to test the client against the tigerblood server
