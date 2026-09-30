-- GET /supplements; same filter defaults as /catalog; pagination enabled
USE maz_suplementos;
SELECT s.id,s.name,s.brand,s.description,s.price,s.stock,s.presentation,s.flavor,s.weight,s.image_url,s.active,s.created_at,s.updated_at
FROM supplements s WHERE (${search}='' OR s.name LIKE CONCAT('%',${search},'%') OR s.brand LIKE CONCAT('%',${search},'%')) ORDER BY s.name;
