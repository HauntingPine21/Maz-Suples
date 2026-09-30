-- DELETE /ingredients/item; id INTEGER query parameter
USE maz_suplementos; UPDATE ingredients SET active=FALSE WHERE id=${id};
