-- PUT /ingredients/item/{id}; id INTEGER path; name, description STRING; active BOOLEAN
USE maz_suplementos; UPDATE ingredients SET name=${name},description=${description},active=${active} WHERE id=${id}; SELECT * FROM ingredients WHERE id=${id};
