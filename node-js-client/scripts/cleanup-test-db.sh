#!/bin/sh

POSTGRES_IP4R_CONTAINER_ID=$(docker ps --filter 'ancestor=postgres-ip4r' -q)

if [ -z "$POSTGRES_IP4R_CONTAINER_ID" ]; then
    echo "Could not find a running postgres-ip4r container."
    exit 1
else
    echo "Found postgres-ip4r container with id: $POSTGRES_IP4R_CONTAINER_ID"
fi

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U postgres tigerblood -c 'DROP TABLE IF EXISTS reputation;'"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U postgres tigerblood -c 'DROP TABLE IF EXISTS violation_reputation_weights;'"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "dropdb -U postgres tigerblood"

docker exec -it $POSTGRES_IP4R_CONTAINER_ID \
       bash -c "psql -U postgres -c 'DROP ROLE IF EXISTS tigerblood;'"
