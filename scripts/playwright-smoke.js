const fs = require("node:fs");
const path = require("node:path");
const { chromium } = require("playwright");

const baseURL = process.env.TARGET_URL || "http://127.0.0.1:8181";
const artifacts =
  process.env.PW_ARTIFACT_DIR || path.join(process.cwd(), "test-artifacts");
fs.mkdirSync(artifacts, { recursive: true });

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
    goals: [{ name: "Fuerza" }],
    ingredients: [{ name: "Creatina monohidratada" }],
  },
  {
    id: 2,
    name: "Maz Whey Vanilla 2 lb",
    brand: "Maz Demo",
    price: 799,
    stock: 12,
    description: "Proteína de suero sabor vainilla.",
    presentation: "Bolsa",
    flavor: "Vainilla",
    weight: "2 lb",
  },
  {
    id: 3,
    name: "Maz Pre-Workout Citrus",
    brand: "Maz Demo",
    price: 529,
    stock: 0,
    description: "Pre-entreno demostrativo.",
    presentation: "Bote",
    flavor: "Cítricos",
    weight: "250 g",
  },
];

(async () => {
  const browser = await chromium.launch({
    headless: process.env.PW_HEADLESS !== "false",
  });
  try {
    const page = await browser.newPage({
      viewport: { width: 1440, height: 900 },
    });
    await page.route("**/api/auth/me", (route) =>
      route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "Sin sesión de prueba" } }),
      }),
    );
    await page.route("**/api/catalog**", async (route) => {
      const url = new URL(route.request().url());
      let body = products;
      if (
        url.pathname.endsWith("/categories") ||
        url.searchParams.get("view") === "categories"
      )
        body = [
          { id: 1, name: "Creatina" },
          { id: 2, name: "Proteína" },
          { id: 3, name: "Pre-entreno" },
        ];
      else if (/\/catalog\/\d+$/.test(url.pathname))
        body = products.find(
          (p) => String(p.id) === url.pathname.split("/").at(-1),
        );
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(body),
      });
    });
    await page.goto(baseURL);
    await page
      .getByRole("heading", { name: "Tu siguiente nivel empieza aquí." })
      .waitFor();
    await page
      .getByRole("heading", { name: "Productos para cada objetivo" })
      .waitFor();
    await page
      .getByRole("heading", { name: "Maz Creatine 300 g", exact: true })
      .waitFor();
    await page.getByRole("button", { name: "Añadir" }).first().click();
    await page.getByRole("button", { name: /Carrito 1/ }).click();
    await page.getByRole("heading", { name: "Tu carrito" }).waitFor();
    const quantity = page.getByRole("spinbutton", { name: /Cantidad de Maz Creatine/ });
    await quantity.fill("2");
    await quantity.press("Tab");
    if (!(await quantity.evaluate((node) => node === globalThis.document.activeElement))) throw new Error("El foco se perdió al cambiar la cantidad del carrito");
    await page.screenshot({
      path: path.join(artifacts, "store-desktop.png"),
      fullPage: true,
    });
    await page.getByRole("button", { name: "Cerrar carrito" }).click();
    await page.getByRole("button", { name: "Ver producto" }).first().click();
    await page.getByRole("dialog", { name: "Maz Creatine 300 g" }).waitFor();
    await page.getByRole("button", { name: "Cerrar detalle" }).click();
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(baseURL);
    const overflow = await page.evaluate(
      () =>
        globalThis.document.documentElement.scrollWidth >
        globalThis.document.documentElement.clientWidth,
    );
    if (overflow)
      throw new Error("La portada tiene desplazamiento horizontal en móvil");
    await page.getByRole("button", { name: "Menú" }).click();
    await page
      .getByRole("navigation", { name: "Navegación principal" })
      .waitFor();
    await page.screenshot({
      path: path.join(artifacts, "store-mobile.png"),
      fullPage: true,
    });
    await page.goto(`${baseURL}/login.html`);
    await page.getByLabel("Usuario").fill("auditor-demo");
    await page.getByLabel("Contraseña").fill("no-se-envia");
    await page.screenshot({
      path: path.join(artifacts, "login-mobile.png"),
      fullPage: true,
    });
    console.log(
      JSON.stringify({
        ok: true,
        title: await page.title(),
        overflow,
        artifacts,
      }),
    );
  } finally {
    await browser.close();
  }
})().catch((error) => {
  console.error(error);
  process.exit(1);
});
