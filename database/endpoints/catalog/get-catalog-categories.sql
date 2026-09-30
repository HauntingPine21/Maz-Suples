-- GET /catalog/categories; pagination enabled
USE maz_suplementos;
SELECT id,name,description FROM categories WHERE active=TRUE ORDER BY name;
