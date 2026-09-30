-- DELETE /supplements/item/{id}; id INTEGER path (desactivación lógica)
USE maz_suplementos;
UPDATE supplements SET active=FALSE WHERE id=${id};
