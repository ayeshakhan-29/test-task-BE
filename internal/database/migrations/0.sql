-- Drop existing schema_migrations table if it exists
DROP TABLE IF EXISTS schema_migrations;

-- Create new schema_migrations table with proper schema
CREATE TABLE schema_migrations (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    version VARCHAR(20) NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY ux_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
