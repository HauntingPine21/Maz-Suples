-- GET /users/item; id INTEGER query parameter
USE maz_suplementos;
SELECT id,username,db_username,full_name,role,active,created_at,updated_at FROM users WHERE id=${id};
