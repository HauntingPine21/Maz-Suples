import { api, el, money, safeImage } from "./api.js";
import * as cart from "./cart.js";
const productsEl = document.querySelector("[data-products]"),
  statusEl = document.querySelector("[data-status]"),
  filterForm = document.querySelector("[data-filters]");
const cartDialog = document.querySelector("[data-cart-dialog]"),
  productDialog = document.querySelector("[data-product-dialog]"),
  checkoutDialog = document.querySelector("[data-checkout-dialog]");
document.querySelector(".nav-toggle")?.addEventListener("click", (event) => {
  const nav = document.querySelector("#nav");
  const open = nav.classList.toggle("open");
  event.currentTarget.setAttribute("aria-expanded", String(open));
});
document
  .querySelectorAll("[data-open-cart]")
  .forEach((b) => b.addEventListener("click", openCart));
document
  .querySelector("[data-close-cart]")
  ?.addEventListener("click", () => cartDialog.close());
document
  .querySelector("[data-close-product]")
  ?.addEventListener("click", () => productDialog.close());
document
  .querySelector("[data-close-checkout]")
  ?.addEventListener("click", () => checkoutDialog.close());
filterForm?.addEventListener("submit", (e) => {
  e.preventDefault();
  loadProducts(new URLSearchParams(new FormData(filterForm)));
});
async function loadCategories() {
  try {
    const rows = await api("/api/catalog?view=categories");
    const select = filterForm.elements.category;
    const strip = document.querySelector("[data-categories]");
    strip.replaceChildren();
    const cats = rows.filter((r) => r.category_name || r.name).slice(0, 8);
    for (const row of cats) {
      const name = String(row.category_name || row.name);
      const option = el("option", { text: name, attrs: { value: name } });
      select.append(option);
      const button = el("button", {
        className: "category-chip",
        text: name,
        attrs: { type: "button" },
      });
      button.addEventListener("click", () => {
        select.value = name;
        filterForm.requestSubmit();
        document.querySelector("#productos").scrollIntoView();
      });
      strip.append(button);
    }
    if (!cats.length)
      strip.append(
        el("p", {
          className: "empty",
          text: "Las categorías aparecerán aquí.",
        }),
      );
  } catch {
    document.querySelector("[data-categories]").textContent =
      "No fue posible cargar las categorías.";
  }
}
async function loadProducts(params = new URLSearchParams()) {
  statusEl.className = "status";
  statusEl.textContent = "Cargando productos…";
  productsEl.replaceChildren();
  for (const [k, v] of [...params]) if (!v) params.delete(k);
  try {
    const rows = await api(`/api/catalog?${params}`);
    const products = rows.filter((r) => r.id && r.name);
    if (!products.length) {
      statusEl.textContent = "No se encontraron productos.";
      return;
    }
    statusEl.textContent = `${products.length} productos encontrados`;
    for (const p of products) productsEl.append(productCard(p));
  } catch (error) {
    statusEl.className = "status error";
    statusEl.textContent = error.message;
  }
}
function productCard(p) {
  const article = el("article", { className: "product-card" });
  const image = el("img", {
    attrs: {
      src: safeImage(p.image_url),
      alt: `${p.name} de ${p.brand || "Maz"}`,
      width: "640",
      height: "640",
      loading: "lazy",
      decoding: "async",
    },
  });
  image.addEventListener(
    "error",
    () => (image.src = "/assets/product-placeholder.svg"),
    { once: true },
  );
  const body = el("div", { className: "product-card-body" });
  body.append(
    el("p", { className: "brand-name", text: String(p.brand || "MAZ") }),
    el("h3", { text: String(p.name) }),
    el("div", { className: "price", text: money.format(Number(p.price) || 0) }),
  );
  const stock = Number(p.stock) || 0;
  body.append(
    el("span", {
      className: `stock ${stock < 1 ? "out" : ""}`,
      text: stock > 0 ? `${stock} disponibles` : "Agotado",
    }),
  );
  const actions = el("div", { className: "product-actions" });
  const detail = el("button", {
    className: "button ghost",
    text: "Ver producto",
    attrs: { type: "button" },
  });
  detail.addEventListener("click", () => showProduct(p.id));
  const add = el("button", {
    className: "button primary",
    text: "Añadir",
    attrs: { type: "button" },
  });
  add.disabled = stock < 1;
  add.addEventListener("click", () => {
    cart.add(p);
    statusEl.className = "status success";
    statusEl.textContent = `${p.name} se añadió al carrito`;
  });
  actions.append(detail, add);
  body.append(actions);
  article.append(image, body);
  return article;
}
async function showProduct(id) {
  const target = document.querySelector("[data-product-detail]");
  const dialogTitle = document.querySelector("#product-title");
  dialogTitle.textContent = "Detalle del producto";
  target.replaceChildren(el("p", { text: "Cargando detalle…" }));
  productDialog.showModal();
  try {
    const p = await api(`/api/catalog/${id}`);
    dialogTitle.textContent = String(p.name);
    const wrap = el("div", { className: "product-detail" });
    wrap.append(
      el("img", {
        attrs: {
          src: safeImage(p.image_url),
          alt: `${p.name} de ${p.brand}`,
          width: "640",
          height: "640",
        },
      }),
    );
    const info = el("div");
    info.append(
      el("p", { className: "eyebrow", text: String(p.brand || "MAZ") }),
      el("h2", { text: String(p.name) }),
      el("p", { className: "price", text: money.format(Number(p.price) || 0) }),
      el("p", { text: String(p.description || "Sin descripción disponible.") }),
      el("p", {
        text: `Presentación: ${p.presentation || "—"} · Sabor: ${p.flavor || "—"} · Contenido: ${p.weight || "—"}`,
      }),
    );
    for (const field of ["categories", "goals", "ingredients"]) {
      const list = parseList(p[field]);
      if (list.length) {
        const tags = el("div", { className: "tag-list" });
        list.forEach((v) =>
          tags.append(
            el("span", { className: "tag", text: String(v.name || v) }),
          ),
        );
        info.append(tags);
      }
    }
    const button = el("button", {
      className: "button primary",
      text: "Añadir al carrito",
      attrs: { type: "button" },
    });
    button.disabled = Number(p.stock) < 1;
    button.addEventListener("click", () => {
      cart.add(p);
      productDialog.close();
      openCart();
    });
    info.append(button);
    wrap.append(info);
    target.replaceChildren(wrap);
  } catch (error) {
    target.textContent = error.message;
  }
}
function parseList(v) {
  if (Array.isArray(v)) return v;
  try {
    return JSON.parse(v || "[]");
  } catch {
    return [];
  }
}
function openCart() {
  cart.render(document.querySelector("[data-cart-items]"));
  document.querySelector("[data-cart-total]").textContent = money.format(
    cart.total(),
  );
  cartDialog.showModal();
}
document.querySelector("[data-checkout]")?.addEventListener("click", () => {
  if (!cart.list().length) return;
  cartDialog.close();
  const summary = document.querySelector("[data-checkout-summary]");
  summary.textContent = `${cart.list().length} productos · Total ${money.format(cart.total())}`;
  checkoutDialog.showModal();
});
let checkoutIdempotencyKey = "";
document
  .querySelector("[data-checkout-form]")
  ?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const state = document.querySelector("[data-checkout-status]");
    state.textContent = "Confirmando pedido…";
    const body = Object.fromEntries(new FormData(e.currentTarget));
    if (!checkoutIdempotencyKey) checkoutIdempotencyKey = crypto.randomUUID();
    body.idempotency_key = checkoutIdempotencyKey;
    body.items = cart
      .list()
      .map((i) => ({ supplement_id: i.id, quantity: i.quantity }));
    try {
      const result = await api("/api/orders", { method: "POST", body });
      const order = Array.isArray(result) ? result[0] : result;
      cart.clear();
      checkoutIdempotencyKey = "";
      state.className = "status success";
      state.textContent = `Pedido confirmado: ${order?.order_number || "consulta el panel"}. Guarda este número.`;
      e.currentTarget.reset();
    } catch (error) {
      state.className = "status error";
      state.textContent = error.message;
    }
  });
loadCategories();
loadProducts();
