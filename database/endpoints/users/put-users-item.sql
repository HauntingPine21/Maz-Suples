-- PUT /users/item/{id}; id INTEGER path; username, full_name, role, password_hash STRING; active BOOLEAN
USE maz_suplementos;
START TRANSACTION PESSIMISTIC;
UPDATE users SET username=${username},full_name=${full_name},role=${role},active=${active},password_hash=IF(${password_hash}='',password_hash,${password_hash}) WHERE id=${id};
DELETE FROM sessions WHERE user_id=${id} AND ${password_hash}<>'';
COMMIT;
SELECT id,username,full_name,role,active,created_at,updated_at FROM users WHERE id=${id};
