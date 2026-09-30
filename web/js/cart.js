import { el, money, safeImage } from "./api.js";
const key = "maz_cart_v1";
let items = load();
function load() {
  try {
    const data = JSON.parse(localStorage.getItem(key) || "[]");
    return Array.isArray(data)
      ? data.filter(
          (i) =>
            Number.isInteger(i.id) &&
            i.id > 0 &&
            Number.isInteger(i.quantity) &&
            i.quantity > 0,
        )
      : [];
  } catch {
    return [];
  }
}
function save() {
  localStorage.setItem(key, JSON.stringify(items));
  document
    .querySelectorAll("[data-cart-count]")
    .forEach(
      (n) =>
        (n.textContent = String(items.reduce((a, i) => a + i.quantity, 0))),
    );
}
export function add(product, quantity = 1) {
  const id = Number(product.id);
  const current = items.find((i) => i.id === id);
  const max = Number(product.stock) || 0;
  if (current) current.quantity = Math.min(current.quantity + quantity, max);
  else
    items.push({
      id,
      name: String(product.name),
      price: Number(product.price),
      stock: max,
      image_url: safeImage(product.image_url),
      quantity: Math.min(quantity, max),
    });
  save();
}
export function list() {
  return items.map((i) => ({ ...i }));
}
export function clear() {
  items = [];
  save();
}
export function total() {
  return items.reduce((a, i) => a + i.price * i.quantity, 0);
}
export function render(container) {
  container.replaceChildren();
  if (!items.length) {
    container.append(
      el("p", { className: "empty", text: "Tu carrito está vacío." }),
    );
    save();
    return;
  }
  for (const [index, item] of items.entries()) {
    const row = el("div", { className: "cart-row" });
    const img = el("img", {
      attrs: { src: item.image_url, alt: "", width: "60", height: "60" },
    });
    const info = el("div");
    info.append(
      el("strong", { text: item.name }),
      el("div", { text: money.format(item.price) }),
    );
    const controls = el("div");
    const input = el("input", {
      attrs: {
        type: "number",
        min: "1",
        max: String(item.stock),
        value: String(item.quantity),
        "aria-label": `Cantidad de ${item.name}`,
      },
    });
    input.addEventListener("change", () => {
      item.quantity = Math.max(
        1,
        Math.min(Number(input.value) || 1, item.stock),
      );
      save();
      render(container);
      container.querySelectorAll('input[type="number"]')[index]?.focus();
    });
    const remove = el("button", {
      text: "Eliminar",
      attrs: { type: "button" },
    });
    remove.addEventListener("click", () => {
      items = items.filter((i) => i.id !== item.id);
      save();
      render(container);
      const removeButtons = container.querySelectorAll(".cart-row button");
      (removeButtons[Math.min(index, removeButtons.length - 1)] ||
        document.querySelector("[data-close-cart]"))?.focus();
    });
    controls.append(input, remove);
    row.append(img, info, controls);
    container.append(row);
  }
  save();
}
save();
