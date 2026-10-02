const { chromium } = require("playwright");
const AxeBuilder = require("@axe-core/playwright").default;

const baseURL = process.env.TARGET_URL || "http://127.0.0.1:8181";
const products = [
  {
    id: 1,
    name: "Maz Creatine 300 g",
    brand: "Maz Demo",
    price: 449,
    stock: 18,
    description: "Creatina monohidratada en polvo.",
    presentation: "Bote",
    flavor: "Sin sabor",
    weight: "300 g",
    categories: [{ name: "Creatina" }],
  },
];

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await context.newPage();
  const findings = [];

  await page.route("**/api/auth/me", (route) =>
    route.fulfill({
      status: 401,
      contentType: "application/json",
      body: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "Sin sesión de prueba" } }),
    }),
  );
  await page.route("**/api/catalog**", (route) => {
    const url = new URL(route.request().url());
    let body = products;
    if (url.pathname.endsWith("/categories") || url.searchParams.get("view") === "categories") {
      body = [{ id: 1, name: "Creatina" }];
    } else if (/\/catalog\/\d+$/.test(url.pathname)) {
      body = products[0];
    }
    return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(body) });
  });

  async function scan(name) {
    const report = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"])
      .analyze();
    for (const violation of report.violations) {
      findings.push({
        name,
        id: violation.id,
        impact: violation.impact,
        nodes: violation.nodes.length,
        targets: violation.nodes.map((node) => node.target),
        summaries: violation.nodes.map((node) => node.failureSummary),
      });
    }
  }

  try {
    await page.goto(baseURL);
    await page.locator(".product-card").waitFor();
    await scan("storefront");
    await page.getByRole("button", { name: "Ver producto" }).click();
    await scan("product-dialog");
    await page.getByRole("button", { name: "Cerrar detalle" }).click();
    await page.getByRole("button", { name: "Añadir", exact: true }).click();
    await page.getByRole("button", { name: /Carrito 1/ }).click();
    await scan("cart-dialog");
    await page.getByRole("button", { name: "Continuar pedido" }).click();
    await scan("checkout-dialog");
    await page.goto(`${baseURL}/login.html`);
    await scan("login");
  } finally {
    await browser.close();
  }

  const blocking = findings.filter(({ impact }) => impact === "serious" || impact === "critical");
  console.log(JSON.stringify({ scans: 5, findings, blocking: blocking.length }));
  if (blocking.length) process.exit(1);
})().catch((error) => {
  console.error(error);
  process.exit(1);
});
