-- POST /orders; id, order_number, customer_name, customer_phone, customer_email, items_json STRING required
-- items_json shape: [{"supplement_id":1,"quantity":2}]
USE maz_suplementos;
START TRANSACTION PESSIMISTIC;
SELECT s.id FROM supplements s JOIN JSON_TABLE(${items_json},'$[*]' COLUMNS(supplement_id BIGINT PATH '$.supplement_id',quantity INT PATH '$.quantity')) j ON j.supplement_id=s.id FOR UPDATE;
INSERT INTO orders(id,order_number,customer_name,customer_phone,customer_email,status,subtotal,total)
SELECT ${id},${order_number},${customer_name},${customer_phone},NULLIF(${customer_email},''),'PENDIENTE',SUM(s.price*j.quantity),SUM(s.price*j.quantity)
FROM JSON_TABLE(${items_json},'$[*]' COLUMNS(supplement_id BIGINT PATH '$.supplement_id',quantity INT PATH '$.quantity')) j JOIN supplements s ON s.id=j.supplement_id
HAVING COUNT(*)=(SELECT COUNT(*) FROM JSON_TABLE(${items_json},'$[*]' COLUMNS(supplement_id BIGINT PATH '$.supplement_id',quantity INT PATH '$.quantity')) x) AND MIN(s.active=TRUE AND s.stock>=j.quantity)=1;
INSERT INTO order_items(order_id,supplement_id,product_name_snapshot,price_snapshot,quantity,subtotal)
SELECT ${id},s.id,s.name,s.price,j.quantity,s.price*j.quantity FROM JSON_TABLE(${items_json},'$[*]' COLUMNS(supplement_id BIGINT PATH '$.supplement_id',quantity INT PATH '$.quantity')) j JOIN supplements s ON s.id=j.supplement_id;
UPDATE supplements s JOIN JSON_TABLE(${items_json},'$[*]' COLUMNS(supplement_id BIGINT PATH '$.supplement_id',quantity INT PATH '$.quantity')) j ON j.supplement_id=s.id SET s.stock=s.stock-j.quantity;
COMMIT;
SELECT id,order_number,status,subtotal,total,created_at FROM orders WHERE id=${id};
