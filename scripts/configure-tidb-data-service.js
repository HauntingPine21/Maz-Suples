const fs = require("fs");
const path = require("path");

module.exports = async function configure({ chromium }) {
  const orgID = process.env.TIDB_CONSOLE_ORG_ID;
  const projectID = process.env.TIDB_CONSOLE_PROJECT_ID;
  const appID = process.env.TIDB_CONSOLE_APP_ID;
  const clusterID = process.env.TIDB_CONSOLE_CLUSTER_ID;
  const controlPlane = process.env.TIDB_CONSOLE_CONTROL_PLANE;
  const cdpURL = process.env.TIDB_CONSOLE_CDP_URL || "http://127.0.0.1:9222";

  if (![orgID, projectID, appID, clusterID, controlPlane].every(Boolean)) {
    throw new Error("Faltan variables TIDB_CONSOLE_* para configurar Data Service");
  }

  const root = path.resolve("database", "endpoints");
  const files = [];
  const walk = (dir) => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) walk(full);
      else if (entry.isFile() && entry.name.endsWith(".sql")) files.push(full);
    }
  };
  walk(root);

  const browser = await chromium.connectOverCDP(cdpURL);
  try {
    const page = browser.contexts()[0].pages().find((candidate) =>
      candidate.url().includes("tidbcloud.com/project/data-service"),
    );
    if (!page) throw new Error("No hay una pestaña abierta de TiDB Cloud Data Service");

    const apiBase = `${controlPlane}/api/v1/dataservices/central/orgs/${orgID}/projects/${projectID}/apps/${appID}/endpoints`;
    const request = async (url, options = {}) => {
      return page.evaluate(async ({ url, options }) => {
        const response = await fetch(url, {
          credentials: "include",
          ...options,
          headers: { "content-type": "application/json", ...(options.headers || {}) },
        });
        const text = await response.text();
        const json = text ? JSON.parse(text) : {};
        if (!response.ok || (json.code && json.code !== 200)) {
          throw new Error(`${response.status}: ${json.message || "Error de TiDB Cloud"}`);
        }
        return json.data || json;
      }, { url, options });
    };

    const listed = await request(apiBase);
    const existing = Array.isArray(listed) ? listed : listed.endpoints || [];
    const byKey = new Map(existing.map((item) => [
      `${item.method}:${item.endpoint || item.name}`,
      item,
    ]));

    const typeFor = (name, route) => {
      if (name === "category_ids") return "array";
      if (name === "active") return "boolean";
      if (["price", "min_price", "max_price"].includes(name)) return "number";
      if (["user_id", "stock"].includes(name)) return "integer";
      if (name === "id" && !route.startsWith("/orders")) return "integer";
      return "string";
    };

    const defaultFor = (name, type) => {
      if (name === "max_price") return "99999999";
      if (["search", "category", "brand", "in_stock", "status", "password_hash"].includes(name)) return "";
      if (type === "number" || type === "integer") return "0";
      if (type === "boolean") return "false";
      if (type === "array") return "";
      return "";
    };

    const retired = new Set([
      "GET:/goals",
      "POST:/goals",
      "PUT:/goals/item",
      "DELETE:/goals/item",
      "GET:/ingredients",
      "POST:/ingredients",
      "PUT:/ingredients/item",
      "DELETE:/ingredients/item",
      "GET:/backup/goals",
      "GET:/backup/ingredients",
      "GET:/backup/supplement_goals",
      "GET:/backup/supplement_ingredients",
	  "GET:/backup/users",
	  "GET:/backup/supplements",
	  "GET:/backup/categories",
	  "GET:/backup/supplement_categories",
	  "GET:/backup/orders",
	  "GET:/backup/order_items",
    ]);
    const configured = [];
    for (const file of files.sort()) {
      const sql = fs.readFileSync(file, "utf8").replace(/\r\n/g, "\n");
      const header = sql.split("\n", 1)[0];
      const match = header.match(/^--\s+(GET|POST|PUT|DELETE)\s+(\/[^;\s]+)/);
      if (!match) throw new Error(`Encabezado inválido: ${file}`);
      const [, method, route] = match;
      if (route === "/users/bootstrap") continue;
      const placeholders = [...new Set([...sql.matchAll(/\$\{([A-Za-z_][A-Za-z0-9_]*)\}/g)].map((m) => m[1]))];
      const args = placeholders.map((name) => {
        const type = typeFor(name, route);
        return {
          name,
          type,
          required: 0,
          default: defaultFor(name, type),
          description: "",
          enum: "",
          item_type: type === "array" ? "integer" : "",
          is_path_parameter: false,
        };
      });
      const pagination = route === "/catalog" || /pagination enabled/i.test(header);
      const rowLimit = /max rows 2000/i.test(header) ? 2000 : 1000;
      const key = `${method}:${route}`;
      let summary = byKey.get(key);

      if (!summary) {
        const created = await request(apiBase, {
          method: "POST",
          body: JSON.stringify({
            name: route,
            sql_template: sql,
            method,
            settings: { row_limit: rowLimit, timeout: 30000 },
            cluster_id: clusterID,
          }),
        });
        summary = created.endpoint || created;
      }

      const current = await request(`${apiBase}/${summary.id}`);
      const payload = {
        ...current,
        id: summary.id,
        dataappp_id: Number(appID),
        name: route,
        endpoint: route,
        method,
        cluster_id: clusterID,
        database: "",
        sql_template: sql,
        settings: {
          timeout: 30000,
          row_limit: rowLimit,
          cache_enabled: 0,
          cache_ttl: 0,
          enable_pagination: pagination ? 1 : 0,
        },
        args,
        tag: "Default",
        batch: false,
        has_modified: true,
      };
      delete payload.session_id;
	  try {
		await request(`${apiBase}/${summary.id}`, {
		  method: "PUT",
		  body: JSON.stringify(payload),
		});
	  } catch (error) {
		throw new Error(`No se pudo configurar ${method} ${route}: ${error.message}`, { cause: error });
	  }
      configured.push({ id: summary.id, method, route, args: args.length, pagination });
    }

    // Retire obsolete endpoints only after every surviving endpoint has been
    // updated successfully, so a partial synchronization cannot remove the
    // old API before the replacement configuration is ready.
    const removed = [];
    for (const key of retired) {
      const summary = byKey.get(key);
      if (!summary) continue;
      await request(`${apiBase}/${summary.id}`, { method: "DELETE" });
      removed.push({ id: summary.id, key });
    }

    console.log(JSON.stringify({ configured: configured.length, removed, endpoints: configured }, null, 2));
  } finally {
    await browser.close();
  }
};
