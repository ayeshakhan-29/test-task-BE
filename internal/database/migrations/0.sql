-- Drop existing schema_migrations table if it exists
DROP TABLE IF EXISTS schema_migrations;

-- Create new schema_migrations table with proper schema
CREATE TABLE schema_migrations (
    version VARCHAR(255) NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
