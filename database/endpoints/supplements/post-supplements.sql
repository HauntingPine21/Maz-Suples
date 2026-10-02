-- POST /supplements; fields from form; category_ids ARRAY(INTEGER), default 0
USE maz_suplementos;
START TRANSACTION;
INSERT INTO supplements(name,brand,description,price,stock,presentation,flavor,weight,image_url,active) VALUES(${name},${brand},${description},${price},${stock},${presentation},NULLIF(${flavor},''),${weight},NULLIF(${image_url},''),${active});
SET @supplement_id=LAST_INSERT_ID();
INSERT IGNORE INTO supplement_categories SELECT @supplement_id,id FROM categories WHERE id IN (${category_ids});
COMMIT;
SELECT * FROM supplements WHERE id=@supplement_id;
