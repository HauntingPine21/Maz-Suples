USE maz_suplementos;

ALTER TABLE users ADD COLUMN db_username VARCHAR(32) NULL AFTER username;

UPDATE users
SET db_username = CASE
  WHEN username REGEXP '^[A-Za-z0-9_.-]{1,32}$' THEN username
  ELSE CONCAT('app_', LEFT(SHA2(LOWER(username), 256), 24))
END
WHERE db_username IS NULL;

ALTER TABLE users MODIFY COLUMN db_username VARCHAR(32) NOT NULL;
ALTER TABLE users ADD UNIQUE KEY uk_users_db_username (db_username);
