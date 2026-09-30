-- GET /orders/item; id STRING query parameter
USE maz_suplementos;
SELECT o.*,COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('supplement_id',oi.supplement_id,'name',oi.product_name_snapshot,'price',oi.price_snapshot,'quantity',oi.quantity,'subtotal',oi.subtotal)) FROM order_items oi WHERE oi.order_id=o.id),'[]') items FROM orders o WHERE o.id=${id};
