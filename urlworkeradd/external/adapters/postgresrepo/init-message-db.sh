#!/bin/bash
set -e
for file in ../migrations/*.sql; do psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"; done