-- GET /catalog/item; id INTEGER query parameter
USE maz_suplementos;
SELECT s.id,s.name,s.brand,s.description,s.price,s.stock,s.presentation,s.flavor,s.weight,s.image_url,
 COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id',c.id,'name',c.name)) FROM supplement_categories sc JOIN categories c ON c.id=sc.category_id WHERE sc.supplement_id=s.id),'[]') categories,
 COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id',g.id,'name',g.name)) FROM supplement_goals sg JOIN goals g ON g.id=sg.goal_id WHERE sg.supplement_id=s.id),'[]') goals,
 COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id',i.id,'name',i.name)) FROM supplement_ingredients si JOIN ingredients i ON i.id=si.ingredient_id WHERE si.supplement_id=s.id),'[]') ingredients
FROM supplements s WHERE s.id=${id} AND s.active=TRUE;
