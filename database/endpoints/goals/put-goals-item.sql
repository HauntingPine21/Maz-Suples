-- PUT /goals/item/{id}; id INTEGER path; name, description STRING; active BOOLEAN
USE maz_suplementos; UPDATE goals SET name=${name},description=${description},active=${active} WHERE id=${id}; SELECT * FROM goals WHERE id=${id};
