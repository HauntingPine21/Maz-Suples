# Política de seguridad

## Versiones compatibles

La rama `master` y el despliegue de producción asociado reciben correcciones de seguridad.

## Reporte responsable

No publiques credenciales, datos personales ni detalles explotables en un issue público. Usa la opción **Report a vulnerability** de la pestaña Security del repositorio:

https://github.com/HauntingPine21/Maz-Suples/security/advisories/new

Incluye el componente afectado, pasos mínimos para reproducir, impacto observado y una forma segura de contacto. El responsable del repositorio confirmará recepción y coordinará la corrección antes de divulgar detalles.

## Secretos

Los secretos de TiDB y Vercel solo deben almacenarse en variables de entorno o administradores de secretos. Ante una exposición, la respuesta obligatoria es rotar primero la credencial y después limpiar el historial.
