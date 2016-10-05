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

## Decay lambda function

In order for the reputation to automatically rise back to 100, you need to set up the lambda function in `./tools/decay/`
