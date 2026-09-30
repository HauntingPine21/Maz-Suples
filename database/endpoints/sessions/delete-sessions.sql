-- DELETE /sessions; token_hash STRING required
USE maz_suplementos;
DELETE FROM sessions WHERE token_hash=${token_hash};
