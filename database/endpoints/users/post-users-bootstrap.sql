-- POST /users/bootstrap; username, db_username, password_hash, full_name STRING. Keep deployed only during bootstrap, then undeploy.
USE maz_suplementos;
INSERT INTO users(username,db_username,password_hash,full_name,role,active)
SELECT ${username},${db_username},${password_hash},${full_name},'ADMINISTRADOR',TRUE WHERE NOT EXISTS (SELECT 1 FROM users WHERE role='ADMINISTRADOR');
SELECT id,username,db_username,full_name,role,active FROM users WHERE username=${username};
