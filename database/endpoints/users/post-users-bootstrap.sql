-- POST /users/bootstrap; username, password_hash, full_name STRING. Keep deployed only during bootstrap, then undeploy.
USE maz_suplementos;
INSERT INTO users(username,password_hash,full_name,role,active)
SELECT ${username},${password_hash},${full_name},'ADMINISTRADOR',TRUE WHERE NOT EXISTS (SELECT 1 FROM users WHERE role='ADMINISTRADOR');
SELECT id,username,full_name,role,active FROM users WHERE username=${username};
