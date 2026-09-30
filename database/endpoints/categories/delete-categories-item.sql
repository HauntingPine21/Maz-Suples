-- DELETE /categories/item/{id}; id INTEGER path (returns FK conflict only if policy changes to physical delete)
USE maz_suplementos; UPDATE categories SET active=FALSE WHERE id=${id};
