#!/bin/bash
# Creates the go-blog database on first cluster initialization.
set -euo pipefail

: "${POSTGRES_USER:?}"

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname postgres <<-EOSQL
	CREATE DATABASE goblog;
EOSQL
