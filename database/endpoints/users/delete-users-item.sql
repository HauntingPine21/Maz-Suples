-- DELETE /users/item/{id}; id INTEGER path (desactivación lógica)
USE maz_suplementos;
UPDATE users SET active=FALSE WHERE id=${id};
