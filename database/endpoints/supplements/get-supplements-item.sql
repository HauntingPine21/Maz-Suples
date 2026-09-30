-- GET /supplements/item/{id}; id INTEGER path
USE maz_suplementos;
SELECT s.*,
 COALESCE((SELECT JSON_ARRAYAGG(category_id) FROM supplement_categories WHERE supplement_id=s.id),'[]') category_ids,
 COALESCE((SELECT JSON_ARRAYAGG(goal_id) FROM supplement_goals WHERE supplement_id=s.id),'[]') goal_ids,
 COALESCE((SELECT JSON_ARRAYAGG(ingredient_id) FROM supplement_ingredients WHERE supplement_id=s.id),'[]') ingredient_ids
FROM supplements s WHERE s.id=${id};
