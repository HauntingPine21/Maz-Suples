-- DELETE /supplements/item; id INTEGER query parameter (desactivación lógica)
USE maz_suplementos;
UPDATE supplements SET active=FALSE WHERE id=${id};
