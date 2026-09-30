-- GET /users/item/{id}; id INTEGER path
USE maz_suplementos;
SELECT id,username,full_name,role,active,created_at,updated_at FROM users WHERE id=${id};
