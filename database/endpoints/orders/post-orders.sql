-- POST /orders; id, order_number, customer_name, customer_phone, customer_email, items_json STRING required
-- items_json shape: [{"supplement_id":1,"quantity":2}]
USE maz_suplementos;
START TRANSACTION;
WITH RECURSIVE item_indexes(idx) AS (
  SELECT 0 WHERE JSON_LENGTH(${items_json})>0
  UNION ALL SELECT idx+1 FROM item_indexes WHERE idx+1<JSON_LENGTH(${items_json}) AND idx<99
), items AS (
  SELECT CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].supplement_id'))) AS UNSIGNED) supplement_id,
         CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].quantity'))) AS UNSIGNED) quantity
  FROM item_indexes
)
SELECT s.id FROM supplements s JOIN items i ON i.supplement_id=s.id FOR UPDATE;
INSERT INTO orders(id,order_number,customer_name,customer_phone,customer_email,status,subtotal,total)
WITH RECURSIVE item_indexes(idx) AS (
  SELECT 0 WHERE JSON_LENGTH(${items_json})>0
  UNION ALL SELECT idx+1 FROM item_indexes WHERE idx+1<JSON_LENGTH(${items_json}) AND idx<99
), items AS (
  SELECT CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].supplement_id'))) AS UNSIGNED) supplement_id,
         CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].quantity'))) AS UNSIGNED) quantity
  FROM item_indexes
)
SELECT ${id},${order_number},${customer_name},${customer_phone},NULLIF(${customer_email},''),'PENDIENTE',SUM(s.price*i.quantity),SUM(s.price*i.quantity)
FROM items i JOIN supplements s ON s.id=i.supplement_id
HAVING COUNT(*)=JSON_LENGTH(${items_json}) AND MIN(s.active=TRUE AND s.stock>=i.quantity)=1;
INSERT INTO order_items(order_id,supplement_id,product_name_snapshot,price_snapshot,quantity,subtotal)
WITH RECURSIVE item_indexes(idx) AS (
  SELECT 0 WHERE JSON_LENGTH(${items_json})>0
  UNION ALL SELECT idx+1 FROM item_indexes WHERE idx+1<JSON_LENGTH(${items_json}) AND idx<99
), items AS (
  SELECT CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].supplement_id'))) AS UNSIGNED) supplement_id,
         CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].quantity'))) AS UNSIGNED) quantity
  FROM item_indexes
)
SELECT ${id},s.id,s.name,s.price,i.quantity,s.price*i.quantity FROM items i JOIN supplements s ON s.id=i.supplement_id;
WITH RECURSIVE item_indexes(idx) AS (
  SELECT 0 WHERE JSON_LENGTH(${items_json})>0
  UNION ALL SELECT idx+1 FROM item_indexes WHERE idx+1<JSON_LENGTH(${items_json}) AND idx<99
), items AS (
  SELECT CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].supplement_id'))) AS UNSIGNED) supplement_id,
         CAST(JSON_UNQUOTE(JSON_EXTRACT(${items_json},CONCAT('$[',idx,'].quantity'))) AS UNSIGNED) quantity
  FROM item_indexes
)
UPDATE supplements s JOIN items i ON i.supplement_id=s.id SET s.stock=s.stock-i.quantity;
COMMIT;
SELECT id,order_number,status,subtotal,total,created_at FROM orders WHERE id=${id};
