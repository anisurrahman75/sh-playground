#!/bin/bash

# Database connection parameters
DB_USER="root"
DB_PASS="my-secret-pw"
DB_NAME="mysql"
TABLE_NAME="anisur"
HOST="127.0.0.1"

# Number of rows to insert
NUM_ROWS=200000000  # Adjust for your needs
BATCH_SIZE=1000     # Number of rows to insert in each batch

# Create temporary SQL file
TEMP_FILE="temp_inserts.sql"
> "$TEMP_FILE"  # Clear the file at the start

# Function to create the table
create_table() {
    mysql -u "$DB_USER" -h "$HOST" -p"$DB_PASS" "$DB_NAME" <<EOF
CREATE TABLE IF NOT EXISTS $TABLE_NAME (
    column1 VARCHAR(255),
    column2 VARCHAR(255),
    column3 VARCHAR(255)
);
EOF
}

# Function to generate insert statements
generate_inserts() {
    for ((i=1; i<=NUM_ROWS; i++)); do
        echo "INSERT INTO $TABLE_NAME (column1, column2, column3) VALUES ('value1_$i', 'value2_$i', 'value3_$i');" >> "$TEMP_FILE"

        # Execute batch insert every BATCH_SIZE
        if (( i % BATCH_SIZE == 0 )); then
            echo "Executing batch of $BATCH_SIZE inserts..."
            mysql -u "$DB_USER" -h "$HOST" -p"$DB_PASS" "$DB_NAME" < "$TEMP_FILE"
            > "$TEMP_FILE"  # Clear the file for the next batch
        fi
    done
}

# Create the table
create_table

# Generate and execute inserts
generate_inserts

# Execute any remaining inserts
if [[ -s "$TEMP_FILE" ]]; then
    echo "Executing remaining inserts..."
    mysql -u "$DB_USER" -h "$HOST" -p"$DB_PASS" "$DB_NAME" < "$TEMP_FILE"
fi

# Cleanup
rm "$TEMP_FILE"

echo "Data insertion completed successfully."
