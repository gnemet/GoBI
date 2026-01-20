#!/bin/bash

# =================================================================
# PostgreSQL Restore Script for 'vir' tables
# =================================================================
# This script restores 'vir' tables from a dump file into the target database.
# It renames the schema from the source (demo_db) to the target schema name.
#
# Usage:
#   ./vir_restore.sh
#
# Environment Variables (replace placeholders or set in .target_env):
#   PGHOST_TARGET
#   PGPORT_TARGET
#   PGUSER_TARGET
#   PGPASSWORD_TARGET
#   TARGET_DB_NAME (e.g., gobi01_db)
#   TARGET_SCHEMA_NAME (e.g., gobi)
#   DUMP_FILE (e.g., vir_tables_dump.sql)
# =================================================================

# Exit immediately if a command exits with a non-zero status.
set -e

# Load environment variables from .target_env file if it exists
if [ -f .target_env ]; then
  export $(cat .target_env | grep -v '#' | awk '/=/ {print $1}')
fi

# --- Set default values if not already set ---
PGHOST_TARGET="${PGHOST_TARGET:-localhost}"
PGPORT_TARGET="${PGPORT_TARGET:-5432}"
PGUSER_TARGET="${PGUSER_TARGET:-postgres}"
TARGET_DB_NAME="${TARGET_DB_NAME:-gobi01_db}"
TARGET_SCHEMA_NAME="${TARGET_SCHEMA_NAME:-gobi}" # Default for target schema
DUMP_FILE="${DUMP_FILE:-vir_tables_dump.sql}"

# Check if PGPASSWORD_TARGET is set, and if not, prompt for it.
if [ -z "$PGPASSWORD_TARGET" ]; then
  echo "Enter password for user $PGUSER_TARGET for target DB ($TARGET_DB_NAME):"
  read -s PGPASSWORD_TARGET
  export PGPASSWORD_TARGET
fi

echo "Attempting to restore tables into $TARGET_DB_NAME, renaming schema to '$TARGET_SCHEMA_NAME'..."

# --- Step 1: Create the target schema if it doesn't exist ---
# This creates the *final* schema name (e.g., 'gobi')
echo "Ensuring target schema '$TARGET_SCHEMA_NAME' exists in $TARGET_DB_NAME..."
psql \
  -h "$PGHOST_TARGET" \
  -p "$PGPORT_TARGET" \
  -U "$PGUSER_TARGET" \
  -d "$TARGET_DB_NAME" \
  -c "CREATE SCHEMA IF NOT EXISTS \"$TARGET_SCHEMA_NAME\";"

echo "Target schema '$TARGET_SCHEMA_NAME' ensured."

# --- Step 2: Restore the dump file ---
# The dump file will create objects under the original schema name (demo_db)
echo "Restoring data from $DUMP_FILE into $TARGET_DB_NAME..."
psql \
  -h "$PGHOST_TARGET" \
  -p "$PGPORT_TARGET" \
  -U "$PGUSER_TARGET" \
  -d "$TARGET_DB_NAME" \
  -f "$DUMP_FILE"

echo "Data restore complete."

# --- Step 3: Rename the schema ---
# The dump created 'demo_db', now rename it to the desired target schema name
if [ "$TARGET_SCHEMA_NAME" != "demo_db" ]; then
  echo "Renaming schema 'demo_db' to '$TARGET_SCHEMA_NAME' in $TARGET_DB_NAME..."
  psql \
    -h "$PGHOST_TARGET" \
    -p "$PGPORT_TARGET" \
    -U "$PGUSER_TARGET" \
    -d "$TARGET_DB_NAME" \
    -c "ALTER SCHEMA demo_db RENAME TO \"$TARGET_SCHEMA_NAME\";"
  echo "Schema renamed successfully."
else
  echo "Target schema name is 'demo_db', no rename needed."
fi

echo "Restore process finished."

# Unset password variable for security
unset PGPASSWORD_TARGET
