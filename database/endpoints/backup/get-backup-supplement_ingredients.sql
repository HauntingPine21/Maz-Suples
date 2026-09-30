-- GET /backup/supplement_ingredients; pagination enabled, max rows 2000
USE maz_suplementos; SELECT * FROM supplement_ingredients ORDER BY supplement_id,ingredient_id;
