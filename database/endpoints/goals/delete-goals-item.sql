-- DELETE /goals/item/{id}; id INTEGER path
USE maz_suplementos; UPDATE goals SET active=FALSE WHERE id=${id};
