-- POST /users; username, password_hash, full_name, role STRING; active BOOLEAN
USE maz_suplementos;
INSERT INTO users(username,password_hash,full_name,role,active) VALUES(${username},${password_hash},${full_name},${role},${active});
SELECT id,username,full_name,role,active,created_at,updated_at FROM users WHERE id=LAST_INSERT_ID();
