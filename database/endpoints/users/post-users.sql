-- POST /users; username, db_username, password_hash, full_name, role STRING; active BOOLEAN
USE maz_suplementos;
INSERT INTO users(username,db_username,password_hash,full_name,role,active) VALUES(${username},${db_username},${password_hash},${full_name},${role},${active});
SELECT id,username,db_username,full_name,role,active,created_at,updated_at FROM users WHERE id=LAST_INSERT_ID();
