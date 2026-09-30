# Endpoints de TiDB Cloud Data Service

Cada archivo incluye en comentarios el método, la ruta y los parámetros que deben configurarse en el Data App. Los parámetros usan la sintaxis oficial `${name}`; Data Service los tipa y enlaza, por lo que el SQL no concatena entradas del navegador.

Los endpoints deben desplegarse desde la consola de TiDB Cloud o sincronizarse como configuración del Data App. No se incluye una llamada administrativa inventada. La creación de esquema y seeds se realiza por una conexión SQL temporal con TLS antes de desplegar los endpoints.

`orders/post-orders.sql` encierra la comprobación de stock, pedido, items y descuento dentro de una sola ejecución con transacción pesimista. Debe probarse en la instancia real antes de considerarlo garantizado, porque Data Service solo documenta públicamente la ejecución secuencial de varias sentencias y la respuesta de la última sentencia.
