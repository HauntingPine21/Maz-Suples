const userAgent = process.env.npm_config_user_agent || "";

if (!userAgent.startsWith("npm/")) {
  console.error("Maz-Suplementos usa npm. Ejecuta npm install o npm ci.");
  process.exit(1);
}
