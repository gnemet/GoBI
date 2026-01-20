#!/bin/bash

# =================================================================
# PostgreSQL Dump Script for 'vir' tables
# =================================================================
# This script dumps 'vir' tables from the specified source schema in the source database.
# The output will contain fully qualified table names (e.g., demo_db.vir_table)
# without including the CREATE SCHEMA statement for the source schema.
#
# Usage:
#   ./vir_dump.sh
#
# Environment Variables (replace placeholders or set in .source_env):
#   PGHOST_SOURCE
#   PGPORT_SOURCE
#   PGUSER_SOURCE
#   PGPASSWORD_SOURCE
#   SOURCE_DB_NAME (e.g., zafir_db)
#   SOURCE_SCHEMA_NAME (e.g., demo_db)
#   DUMP_FILE (e.g., vir_tables_dump.sql)
# =================================================================

# Exit immediately if a command exits with a non-zero status.
set -e

# Load environment variables from .source_env file if it exists
if [ -f .source_env ]; then
  export $(cat .source_env | grep -v '#' | awk '/=/ {print $1}')
fi

# --- Set default values if not already set ---
PGHOST_SOURCE="${PGHOST_SOURCE:-localhost}"
PGPORT_SOURCE="${PGPORT_SOURCE:-5432}"
PGUSER_SOURCE="${PGUSER_SOURCE:-postgres}"
SOURCE_DB_NAME="${SOURCE_DB_NAME:-zafir_db}"
SOURCE_SCHEMA_NAME="${SOURCE_SCHEMA_NAME:-demo_db}" # Default for source schema
DUMP_FILE="${DUMP_FILE:-vir_tables_dump.sql}"

# Check if PGPASSWORD_SOURCE is set, and if not, prompt for it.
if [ -z "$PGPASSWORD_SOURCE" ]; then
  echo "Enter password for user $PGUSER_SOURCE for source DB ($SOURCE_DB_NAME):"
  read -s PGPASSWORD_SOURCE
  export PGPASSWORD_SOURCE
fi

echo "Dumping tables from $SOURCE_DB_NAME.$SOURCE_SCHEMA_NAME..."

# Removed -n "$SOURCE_SCHEMA_NAME" so that table names are fully qualified in the dump.
# This allows for easier sed replacement later.
pg_dump \
  -h "$PGHOST_SOURCE" \
  -p "$PGPORT_SOURCE" \
  -U "$PGUSER_SOURCE" \
  -d "$SOURCE_DB_NAME" \
  -t "$SOURCE_SCHEMA_NAME.vir*" \
  --format=plain \
  --no-owner \
  --no-privileges \
  --file="$DUMP_FILE"

echo "Dump complete. Output saved to $DUMP_FILE"

# Unset password variable for security
unset PGPASSWORD_SOURCE
