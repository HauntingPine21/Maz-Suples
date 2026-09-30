-- GET /orders; status STRING='' default; pagination enabled
USE maz_suplementos; SELECT id,order_number,customer_name,customer_phone,customer_email,status,subtotal,total,created_at,updated_at FROM orders WHERE (${status}='' OR status=${status}) ORDER BY created_at DESC;
