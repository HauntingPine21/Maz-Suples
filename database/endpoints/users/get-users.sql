-- GET /users; pagination enabled
USE maz_suplementos;
SELECT id,username,db_username,full_name,role,active,created_at,updated_at FROM users ORDER BY username;
