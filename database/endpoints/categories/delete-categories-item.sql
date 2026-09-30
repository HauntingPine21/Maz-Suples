-- DELETE /categories/item; id INTEGER query parameter
USE maz_suplementos; UPDATE categories SET active=FALSE WHERE id=${id};
