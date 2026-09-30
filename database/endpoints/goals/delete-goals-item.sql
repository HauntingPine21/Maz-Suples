-- DELETE /goals/item; id INTEGER query parameter
USE maz_suplementos; UPDATE goals SET active=FALSE WHERE id=${id};
