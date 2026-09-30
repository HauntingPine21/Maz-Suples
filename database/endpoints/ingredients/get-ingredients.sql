-- GET /ingredients; pagination enabled
USE maz_suplementos; SELECT id,name,description,active,created_at,updated_at FROM ingredients ORDER BY name;
