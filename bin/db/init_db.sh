#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    DROP DATABASE IF EXISTS tigerblood;
    DROP ROLE IF EXISTS tigerblood;

    CREATE ROLE tigerblood WITH LOGIN PASSWORD '$POSTGRES_PASSWORD';
    CREATE DATABASE tigerblood;
    GRANT ALL PRIVILEGES ON DATABASE tigerblood TO tigerblood;
EOSQL

psql -v ON_ERROR_STOP=1 -U postgres tigerblood -c "CREATE EXTENSION IF NOT EXISTS ip4r;"
