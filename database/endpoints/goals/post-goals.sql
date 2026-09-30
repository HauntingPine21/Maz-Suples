-- POST /goals; name, description STRING; active BOOLEAN
USE maz_suplementos; INSERT INTO goals(name,description,active) VALUES(${name},${description},${active}); SELECT * FROM goals WHERE id=LAST_INSERT_ID();
