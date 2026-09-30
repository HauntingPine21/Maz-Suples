-- POST /categories; name, description STRING; active BOOLEAN
USE maz_suplementos; INSERT INTO categories(name,description,active) VALUES(${name},${description},${active}); SELECT * FROM categories WHERE id=LAST_INSERT_ID();
