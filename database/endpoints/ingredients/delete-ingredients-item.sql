-- DELETE /ingredients/item/{id}; id INTEGER path
USE maz_suplementos; UPDATE ingredients SET active=FALSE WHERE id=${id};
