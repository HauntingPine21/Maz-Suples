-- PUT /ingredients/item; id INTEGER body; name, description STRING; active BOOLEAN
USE maz_suplementos; UPDATE ingredients SET name=${name},description=${description},active=${active} WHERE id=${id}; SELECT * FROM ingredients WHERE id=${id};
