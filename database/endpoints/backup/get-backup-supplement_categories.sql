-- GET /backup/supplement_categories; pagination enabled, max rows 2000
USE maz_suplementos; SELECT * FROM supplement_categories ORDER BY supplement_id,category_id;
