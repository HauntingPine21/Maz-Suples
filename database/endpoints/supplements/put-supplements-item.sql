-- PUT /supplements/item; id INTEGER body; remaining params as POST
USE maz_suplementos;
START TRANSACTION;
UPDATE supplements SET name=${name},brand=${brand},description=${description},price=${price},stock=${stock},presentation=${presentation},flavor=NULLIF(${flavor},''),weight=${weight},image_url=NULLIF(${image_url},''),active=${active} WHERE id=${id};
DELETE FROM supplement_categories WHERE supplement_id=${id};
INSERT IGNORE INTO supplement_categories SELECT ${id},id FROM categories WHERE id IN (${category_ids});
COMMIT;
SELECT * FROM supplements WHERE id=${id};
