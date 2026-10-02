-- GET /catalog; params: search STRING='', category STRING='', brand STRING='', min_price NUMBER=0, max_price NUMBER=99999999, in_stock STRING=''
USE maz_suplementos;
SELECT s.id,s.name,s.brand,s.description,s.price,s.stock,s.presentation,s.flavor,s.weight,s.image_url,s.active,s.created_at,s.updated_at
FROM supplements s
WHERE s.active=TRUE
  AND (${search}='' OR s.name LIKE CONCAT('%',${search},'%') OR s.brand LIKE CONCAT('%',${search},'%'))
  AND (${brand}='' OR s.brand=${brand})
  AND s.price BETWEEN ${min_price} AND ${max_price}
  AND (${in_stock}='' OR ${in_stock}<>'true' OR s.stock>0)
  AND (${category}='' OR EXISTS (SELECT 1 FROM supplement_categories sc JOIN categories c ON c.id=sc.category_id WHERE sc.supplement_id=s.id AND c.name=${category}))
ORDER BY s.name;
