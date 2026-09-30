-- GET /goals; pagination enabled
USE maz_suplementos; SELECT id,name,description,active,created_at,updated_at FROM goals ORDER BY name;
