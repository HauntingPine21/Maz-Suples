-- GET /supplements/item; id INTEGER query parameter
USE maz_suplementos;
SELECT s.*,
 COALESCE((SELECT JSON_ARRAYAGG(category_id) FROM supplement_categories WHERE supplement_id=s.id),'[]') category_ids
FROM supplements s WHERE s.id=${id};
