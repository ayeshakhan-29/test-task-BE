-- Add full_name column and remove username
ALTER TABLE users
ADD COLUMN full_name VARCHAR(100) NOT NULL AFTER id,
DROP COLUMN username;
