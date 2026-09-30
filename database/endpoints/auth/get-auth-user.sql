-- GET /auth/user; username STRING required
USE maz_suplementos;
SELECT id,username,password_hash,full_name,role,active FROM users WHERE username=${username} LIMIT 1;
