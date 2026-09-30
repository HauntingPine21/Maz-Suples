-- POST /supplements; fields from form; category_ids, goal_ids, ingredient_ids ARRAY(INTEGER), default 0
USE maz_suplementos;
START TRANSACTION PESSIMISTIC;
INSERT INTO supplements(name,brand,description,price,stock,presentation,flavor,weight,image_url,active) VALUES(${name},${brand},${description},${price},${stock},${presentation},NULLIF(${flavor},''),${weight},NULLIF(${image_url},''),${active});
SET @supplement_id=LAST_INSERT_ID();
INSERT IGNORE INTO supplement_categories SELECT @supplement_id,id FROM categories WHERE id IN (${category_ids});
INSERT IGNORE INTO supplement_goals SELECT @supplement_id,id FROM goals WHERE id IN (${goal_ids});
INSERT IGNORE INTO supplement_ingredients SELECT @supplement_id,id FROM ingredients WHERE id IN (${ingredient_ids});
COMMIT;
SELECT * FROM supplements WHERE id=@supplement_id;
