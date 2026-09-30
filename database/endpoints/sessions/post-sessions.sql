-- POST /sessions; token_hash STRING, csrf_hash STRING, user_id INTEGER, expires_at STRING required
USE maz_suplementos;
DELETE FROM sessions WHERE expires_at<=CURRENT_TIMESTAMP;
INSERT INTO sessions(token_hash,csrf_hash,user_id,expires_at) VALUES(${token_hash},${csrf_hash},${user_id},${expires_at});
SELECT user_id FROM sessions WHERE token_hash=${token_hash};
