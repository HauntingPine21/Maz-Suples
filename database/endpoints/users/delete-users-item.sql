-- DELETE /users/item; id INTEGER query parameter (desactivación lógica)
USE maz_suplementos;
UPDATE users SET active=FALSE WHERE id=${id};
