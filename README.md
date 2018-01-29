# Tigerblood

Mozilla's IP-based reputation service.

## Running the tests

In order to run the tests, you need a local postgresql database listening on port 5432 with a user `tigerblood` (without a password) which has access to a database calles `tigerblood`, and with the `ip4r` extension installed and created..

If you don't want to install postgres, you can do this from a docker container:

```console
make build-db start-db
# wait for db to start
make setup-db
```

## Healthcheck

You can find a healthcheck lambda function under `./tools/healthcheck`.

## Configuration

Tigerblood can be configured either via a config file or via environment variables.

The following configuration options are available:

| Option name                | Description                                                                              | Default           |
|----------------------------|------------------------------------------------------------------------------------------|-------------------|
| CREDENTIALS                | A map of hawk id-keys.                                                                   | -                 |
| DATABASE\_MAX\_OPEN\_CONNS | The maximum amount of PostgreSQL database connections tigerblood will open               | 80                |
| BIND\_ADDR                 | The host and port tigerblood will listen on for HTTP requests                            | 127.0.0.1:8080    |
| DSN                        | The PostgreSQL data source name. Mandatory.                                              | -                 |
| HAWK                       | true to enable Hawk authentication. If true is provided, credentials must be non-empty   | false             |
| VIOLATION_PENALTIES        | A map of violation names to their reputation penalty weight 0 to 100 inclusive. Ignores violation names with dashes.          | -                 |
| STATSD\_ADDR               | The host and port for statsd                                                             | 127.0.0.1:8125    |
| STATSD\_NAMESPACE          | The statsd namespace prefix                                                              | tigerblood.       |
| PUBLISH\_RUNTIME\_STATS    | true to enable sending go runtime stats to STATSD\_ADDR                                  | false             |
| RUNTIME\_PAUSE\_INTERVAL   | How often to send go runtime stats in seconds                                            | 10                |
| RUNTIME\_CPU               | Send `cpu.goroutines` and `cpu.cgo_calls` when runtime stats are enabled.                | true              |
| RUNTIME\_MEM               | Send top level `mem`, `mem.heap`, and `mem.stack` stats when runtime stats are enabled.  | true              |
| RUNTIME\_GC                | Send `mem.gc` stats when runtime stats are enabled.                                      | true              |
| MAX_ENTRIES                | Maximum number of entries for multi entry endpoints to accept                            | 1000              |

For environment variables, the configuration options must be prefixed with "TIGERBLOOD\_", for example, the environment variable to configure the DSN is TIGERBLOOD\_DSN.

The config file can be JSON, TOML, YAML, HCL, or a Java properties file. Keys do not have to be prefixed in config files. For example:

```json
{
    "DSN": "user=tigerblood dbname=tigerblood sslmode=disable",
    "BIND_ADDR": "127.0.0.1:8080",
    "HAWK": "yes",
    "CREDENTIALS": {
        "root": "toor"
    },
    "VIOLATION_PENALTIES": "rate_limit_exceeded=2"
}
```

After setting up the db, we can use the example config file to run the service:

```
cp config.yml.example config.yml
make run
```

## Decay lambda function

In order for the reputation to automatically rise back to 100, you need to set up the lambda function in `./tools/decay/`


## HTTP API

Response schema:

```json
{
  "type": "object",
  "properties": {
    "IP": {
      "type": "string"
    },
    "Reputation": {
      "type": "integer"
    }
  },
  "required": [
    "IP",
    "Reputation"
  ]
}
```

### Authorization

All requests to the API must be authenticated with a [Hawk](https://github.com/hueniverse/hawk) authorization header. For example, if you're doing requests with Python's `requests` package, you can use [requests-hawk](https://github.com/mozilla-services/requests-hawk) to generate headers. [The Hawk readme](https://github.com/hueniverse/hawk#implementations) contains information on different implementations for other languages. Request bodies are validated by the server (https://github.com/hueniverse/hawk#payload-validation), but the server does not provide any mechanism for response validation.

### Endpoints
`{ip}` should be substituted for a CIDR-notation IP address or network.
In the examples, we assume tigerblood is listening on http://tigerblood

#### GET /{ip}

Retrieves information about an IP address or network.

* Request body: None
* Request parameters: None

* Response body: a JSON object with the schema specified above
* Successful response status code: 200

Example: `curl http://tigerblood/240.0.0.1 --header "Authorization: {YOUR_HAWK_HEADER}"`

#### POST /

Records information about a new IP address or network.

* Request body: a JSON object with the schema specified above
* Request parameters: None

* Response body: None
* Successful response status code: 201

Example: `curl -d '{"IP": "240.0.0.1", "Reputation": 45}' -X POST http://tigerblood/ --header "Authorization: {YOUR_HAWK_HEADER}"`

#### PUT /{ip}

Updates information about an IP address or network.

* Request body: a JSON object with the schema specified above. The `"IP"` field is optional for this endpoint, and if provided, it will be ignored.

* Response body: None
* Successful response status code: 200

Example: `curl -d '{"Reputation": 5}' -X PUT http://tigerblood/240.0.0.1 --header "Authorization: {YOUR_HAWK_HEADER}"`

#### DELETE /{ip}

Deletes information about an IP address or network.

* Request body: None
* Request parameters: None

* Response body: None
* Successful response status code: 200

Example: `curl -X DELETE http://tigerblood/240.0.0.1 --header "Authorization: {YOUR_HAWK_HEADER}"`

#### GET /__lbheartbeat__ and GET /__heartbeat__

Endpoints designed for load balancers.

* Request body: None
* Request parameters: None

* Response body: None
* Successful response status code: 200

Example: `curl http://tigerblood/__heartbeat__`

#### GET /__version__

* Request body: None
* Request parameters: None

* Response body: A JSON object with information about tigerblood's version.
* Successful response status code: 200

Example: `curl http://tigerblood/__version__`

#### GET /violations

* Request parameters: None
* Request body: a JSON object with the schema:

Returns a hashmap of violation type to penalty loaded from the config e.g.

```json
{
  "violationName": 20,
  "testViolation": 0
}
```

* Successful response status code: 200

Example: `curl -X GET http://tigerblood/violations`

#### PUT /violations/{ip}

Sets or updates the reputation for an IP address or network to the
reputation for the violation type found in the config if it is lower
than the current reputation.


* Request parameters: None
* Request body: a JSON object with the schema:

```json
{
  "type": "object",
  "properties": {
    "Violation": {
      "type": "string"
    }
  },
  "required": [
    "Violation"
  ]
}
```

* Response body: None
* Successful response status code: 204 No Content

Example: `curl -d '{"Violation": "password-check-rate-limited-exceeded"}' -X PUT http://tigerblood/violations/240.0.0.1 --header "Authorization: {YOUR_HAWK_HEADER}"`

#### PUT /violations/

Sets or updates the reputations for multiple IP addresses or networks
with provided violation types like `PUT /violations/{ip}` for each IP.

Returns 409 Conflict for requests with duplicate IPs.

In the event of an invalid or failed entry, returns the failing entry
and index with the error response body below and rolls back
the accepted entries to retry (i.e. everything runs as one SQL statement).

Max entries can be configured with the `TIGERBLOOD_MAX_ENTRIES` env var,
which default to 1000.

* Request parameters: None
* Request body:

A JSON object with the schema (example below):

```json
[{
  "type": "object",
  "properties": {
    "Violation": {
      "type": "string"
    },
    "ip": {
      "type": "string"
    }
  },
  "required": [
    "Violation",
	"ip"
  ]
}]
```

* Response body: None
* Successful response status code: 204 No Content

* Error Response body:

A JSON object with the schema (example below):

```
{
  "type": "object",
  "properties": {
    "Errno": {
      "type": "int"
    },
    "EntryIndex": {
      "type": "int"
    },
    "Entry": {
      "type": "object",
      "properties": {
         "Violation": {
           "type": "string"
         },
         "IP": {
           "type": "string"
         }
      }
	},
	"Msg": {
      "type": "string"
	}
  }
}
```

* Error response status code: 400 Bad Request

Example: `curl -d '[{"ip": , "Violation": "password-check-rate-limited-exceeded"}]' -X PUT http://tigerblood/violations/ --header "Authorization: {YOUR_HAWK_HEADER}"`

Example error response: `{\"EntryIndex\":0,\"Entry\":{\"IP\":\"192.168.0.1\",\"Violation\":\"Unknown\"},\"Msg\":\"Violation type not found\"}`

## CLI API

A CLI tool for Ops to ban IPs manually lives in `cmd/tigerblood-cli`.

To install it:

```console
go get -v github.com/mozilla-services/tigerblood/cmd/tigerblood-cli
```

Check that it's working:

```console
tigerblood-cli help
Command line client for managing IP Reputations. It requires
the environment variables TIGERBLOOD_HAWK_ID, TIGERBLOOD_HAWK_SECRET, TIGERBLOOD_URL. Example usage:

TIGERBLOOD_HAWK_ID=root TIGERBLOOD_HAWK_SECRET=toor TIGERBLOOD_URL=http://localhost:8080/ tigerblood-cli ban 192.8.8.0

Usage:
  tigerblood-cli [command]

Available Commands:
  ban         Ban an IP for the maximum decay period (environment dependent).
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.tigerblood-cli.yaml)
  -h, --help            help for tigerblood-cli
  -t, --toggle          Help message for toggle

Use "tigerblood-cli [command] --help" for more information about a command.
```

To ban an IP:

1. Get HAWK creds from the @foxsec team
1. Export them into your environment e.g.

```console
export TIGERBLOOD_HAWK_ID=root
export TIGERBLOOD_HAWK_SECRET=toor
export TIGERBLOOD_URL=http://localhost:8080/
```

1. Ban things temporarily:

```console
tigerblood-cli ban 0.0.0.0
```
