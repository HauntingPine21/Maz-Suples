-- GET /sessions/current; token_hash STRING required
USE maz_suplementos;
SELECT s.user_id,s.csrf_hash,u.username,u.full_name,u.role FROM sessions s JOIN users u ON u.id=s.user_id
WHERE s.token_hash=${token_hash} AND s.expires_at>CURRENT_TIMESTAMP AND u.active=TRUE LIMIT 1;
