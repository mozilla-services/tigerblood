# Tigerblood

Mozilla's IP-based reputation service.

## Running the tests

In order to run the tests, you need a local postgresql database listening on port 5432 with a user `tigerblood` (without a password) which has access to a database calles `tigerblood`, and with the `ip4r` extension installed and created..

If you don't want to install postgres, you can do this from a docker container:

- Build the container with `docker build -f postgres.Dockerfile -t postgres-ip4r .`.
- Run `docker run --name postgres -p 127.0.0.1:5432:5432 -d postgres-ip4r` to create a postgres container bound to port 5432 locally.
- Run `docker exec -ti postgres bash` to get a shell inside the container.
- From this shell, run `psql -U postgres`. You should get a postgres prompt.
- Run the following SQL to create the `tigerblood` user and database:
  ```sql
  
  CREATE ROLE tigerblood WITH LOGIN;
  CREATE DATABASE tigerblood;
  GRANT ALL PRIVILEGES ON DATABASE tigerblood TO tigerblood;
  \c tigerblood
  CREATE EXTENSION ip4r;
  ```
- Exit postgres by typing `\q` and pressing the Return key.
- Exit the docker container by typing `exit` and pressing the Return key.

## HTTP API

All requests to the API must be authenticated with a Hawk authorization header.

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

`{ip}` should be substituted for a CIDR-notation IP address or network.
In the examples, we assume tigerblood is listening on http://tigerblood

### GET /{ip}

Retrieves information about an IP address or network.

* Request body: None
* Request parameters: None

* Response body: a JSON object with the schema specified above
* Successful response status code: 200

Example: `curl http://tigerblood/240.0.0.1`

### POST /

Records information about a new IP address or network.

* Request body: a JSON object with the schema specified above
* Request parameters: None

* Response body: None
* Successful response status code: 201

Example: `curl -d '{"IP": "240.0.0.1", "Reputation": 45}' -X POST http://tigerblood/`

### PUT /{ip}

Updates information about an IP address or network.

* Request body: a JSON object with the schema specified above. The `"IP"` field is optional for this endpoint, and if provided, it will be ignored.

* Response body: None
* Successful response status code: 200

Example: `curl -d '{"Reputation": 5}' -X PUT http://tigerblood/240.0.0.1`

### DELETE /{ip}

Deletes information about an IP address or network.

* Request body: None
* Request parameters: None

* Response body: None
* Successful response status code: 200

Example: `curl -X DELETE http://tigerblood/240.0.0.1`

### GET /__lbheartbeat__ and GET /__heartbeat__

Endpoints designed for load balancers.

* Request body: None
* Request parameters: None

* Response body: None
* Successful response status code: 200

Example: `curl http://tigerblood/__heartbeat__`

### GET /__version__

* Request body: None
* Request parameters: None

* Response body: A JSON object with information about tigerblood's version.
* Successful response status code: 200

Example: `curl http://tigerblood/__version__`
