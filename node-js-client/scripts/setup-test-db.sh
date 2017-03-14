#!/bin/sh

POSTGRES_IP4R_CONTAINER_ID=$(docker ps --filter 'ancestor=postgres-ip4r' -q)

if [ -z "$POSTGRES_IP4R_CONTAINER_ID" ]; then
    echo "Could not find a running postgres-ip4r container."
    exit 1
else
    echo "Found postgres-ip4r container with id: $POSTGRES_IP4R_CONTAINER_ID"
fi

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "createdb -U postgres tigerblood"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U postgres -c 'CREATE ROLE tigerblood WITH LOGIN; GRANT ALL PRIVILEGES ON DATABASE tigerblood TO tigerblood;'"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U postgres tigerblood -c 'CREATE EXTENSION ip4r;'"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U tigerblood tigerblood -c 'CREATE TABLE IF NOT EXISTS violation_reputation_weights (
violation_type varchar(128) PRIMARY KEY NOT NULL,
reputation int NOT NULL CHECK (reputation >= 0 AND reputation <= 100),
UNIQUE (violation_type, reputation)
);'"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U tigerblood tigerblood -c \"INSERT INTO violation_reputation_weights (violation_type, reputation) VALUES ('test_violation', 30);\""
