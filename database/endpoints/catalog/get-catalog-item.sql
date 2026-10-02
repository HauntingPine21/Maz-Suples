-- GET /catalog/item; id INTEGER query parameter
USE maz_suplementos;
SELECT s.id,s.name,s.brand,s.description,s.price,s.stock,s.presentation,s.flavor,s.weight,s.image_url,
 COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id',c.id,'name',c.name)) FROM supplement_categories sc JOIN categories c ON c.id=sc.category_id WHERE sc.supplement_id=s.id),'[]') categories
FROM supplements s WHERE s.id=${id} AND s.active=TRUE;
