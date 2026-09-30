-- PUT /categories/item/{id}; id INTEGER path; name, description STRING; active BOOLEAN
USE maz_suplementos; UPDATE categories SET name=${name},description=${description},active=${active} WHERE id=${id}; SELECT * FROM categories WHERE id=${id};
