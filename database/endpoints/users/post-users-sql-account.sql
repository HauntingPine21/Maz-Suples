-- POST /users/sql_account; db_username, db_password STRING
SET @maz_create_user_sql = CONCAT(
  'CREATE USER IF NOT EXISTS ',
  QUOTE(${db_username}),
  '@''%'' IDENTIFIED BY ',
  QUOTE(${db_password})
);
PREPARE maz_create_user FROM @maz_create_user_sql;
EXECUTE maz_create_user;
DEALLOCATE PREPARE maz_create_user;
SELECT User AS db_username, Host AS db_host
FROM mysql.user
WHERE User=${db_username} AND Host='%';
