-- PUT /orders/status; id STRING body; status STRING enum
USE maz_suplementos; UPDATE orders SET status=${status} WHERE id=${id}; SELECT * FROM orders WHERE id=${id};
