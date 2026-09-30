import { api, dateTime, el, money, setCSRF } from "./api.js";
const content = document.querySelector("[data-admin-content]"),
  status = document.querySelector("[data-admin-status]"),
  title = document.querySelector("[data-view-title]"),
  dialog = document.querySelector("[data-form-dialog]"),
  formStatus = document.querySelector("[data-form-status]"),
  form = document.querySelector("[data-resource-form]");
let user,
  currentView = "dashboard";
const labels = {
  dashboard: "Resumen",
  supplements: "Productos",
  categories: "Categorías",
  goals: "Objetivos",
  ingredients: "Ingredientes",
  orders: "Pedidos",
  users: "Usuarios",
  backups: "Respaldos",
};
const permissions = {
  ADMINISTRADOR: [
    "dashboard",
    "supplements",
    "categories",
    "goals",
    "ingredients",
    "orders",
    "users",
    "backups",
  ],
  CAPTURISTA: [
    "dashboard",
    "supplements",
    "categories",
    "goals",
    "ingredients",
    "orders",
  ],
  AUDITOR: [
    "dashboard",
    "supplements",
    "categories",
    "goals",
    "ingredients",
    "orders",
    "users",
  ],
};
async function init() {
  try {
    const data = await api("/api/auth/me");
    user = data.user;
    if (data.csrf_token) setCSRF(data.csrf_token);
    document.querySelector("[data-user-badge]").textContent =
      `${user.full_name} · ${user.role}`;
    document.querySelectorAll("[data-view]").forEach((button) => {
      if (!permissions[user.role]?.includes(button.dataset.view))
        button.hidden = true;
      button.addEventListener("click", () => show(button.dataset.view));
    });
    document.querySelector("[data-logout]").addEventListener("click", logout);
    document
      .querySelector("[data-close-form]")
      .addEventListener("click", () => dialog.close());
    show("dashboard");
  } catch (error) {
    if (error.status === 401) location.assign("/login.html");
    else setStatus(error.message, true);
  }
}
async function show(view) {
  currentView = view;
  title.textContent = labels[view] || view;
  document
    .querySelectorAll("[data-view]")
    .forEach((b) => {
      const active = b.dataset.view === view;
      b.classList.toggle("active", active);
      b.setAttribute("aria-current", active ? "page" : "false");
    });
  setStatus("Cargando…");
  content.replaceChildren();
  try {
    if (view === "dashboard") await dashboard();
    else if (view === "backups") backups();
    else await listing(view);
    setStatus("");
  } catch (error) {
    if (error.status === 401) location.assign("/login.html");
    else setStatus(error.message, true);
  }
}
async function dashboard() {
  const [products, orders] = await Promise.all([
    api("/api/supplements?page_size=200"),
    api("/api/orders?page_size=200"),
  ]);
  const active = products.filter((p) => truthy(p.active)),
    low = products.filter((p) => Number(p.stock) > 0 && Number(p.stock) <= 5),
    out = products.filter((p) => Number(p.stock) === 0),
    pending = orders.filter((o) => o.status === "PENDIENTE"),
    completed = orders.filter((o) => o.status === "COMPLETADO"),
    value = products.reduce(
      (sum, p) => sum + Number(p.price) * Number(p.stock),
      0,
    );
  const grid = el("div", { className: "kpi-grid" });
  [
    ["Productos activos", active.length],
    ["Stock bajo", low.length],
    ["Agotados", out.length],
    ["Pedidos pendientes", pending.length],
    ["Completados", completed.length],
    ["Valor de inventario", money.format(value)],
  ].forEach(([name, value]) => {
    const card = el("article", { className: "kpi" });
    card.append(
      el("span", { text: name }),
      el("strong", { text: String(value) }),
    );
    grid.append(card);
  });
  content.append(grid);
}
async function listing(resource) {
  const rows = await api(`/api/${resource}?page_size=200`);
  const canWrite =
    user.role === "CAPTURISTA" && resource !== "users" && resource !== "orders";
  const canUsers = user.role === "ADMINISTRADOR" && resource === "users";
  const toolbar = el("div", { className: "table-toolbar" });
  toolbar.append(el("p", { text: `${rows.length} registros` }));
  if (canWrite || canUsers) {
    const add = el("button", {
      className: "button primary",
      text: "Crear nuevo",
      attrs: { type: "button" },
    });
    add.addEventListener("click", () => openForm(resource));
    toolbar.append(add);
  }
  content.append(toolbar);
  if (!rows.length) {
    content.append(
      el("p", { className: "empty", text: "No hay registros para mostrar." }),
    );
    return;
  }
  const columns = columnsFor(resource);
  const wrap = el("div", { className: "table-wrap" }),
    table = el("table"),
    thead = el("thead"),
    head = el("tr");
  columns.forEach((c) =>
    head.append(el("th", { text: c.label, attrs: { scope: "col" } })),
  );
  if (
    canWrite ||
    canUsers ||
    (resource === "orders" && user.role === "CAPTURISTA")
  )
    head.append(el("th", { text: "Acciones", attrs: { scope: "col" } }));
  thead.append(head);
  const tbody = el("tbody");
  for (const row of rows) {
    const tr = el("tr");
    columns.forEach((c) =>
      tr.append(el("td", { text: format(row[c.key], c.type) })),
    );
    if (
      canWrite ||
      canUsers ||
      (resource === "orders" && user.role === "CAPTURISTA")
    ) {
      const cell = el("td");
      const edit = el("button", {
        className: "button ghost",
        text: resource === "orders" ? "Estado" : "Editar",
        attrs: { type: "button" },
      });
      edit.addEventListener("click", () =>
        resource === "orders" ? orderStatus(row) : openForm(resource, row),
      );
      cell.append(edit);
      if (resource !== "orders") {
        const remove = el("button", {
          className: "button ghost",
          text: "Desactivar",
          attrs: { type: "button" },
        });
        remove.addEventListener("click", () => removeRow(resource, row));
        cell.append(remove);
      }
      tr.append(cell);
    }
    tbody.append(tr);
  }
  table.append(thead, tbody);
  wrap.append(table);
  content.append(wrap);
}
function columnsFor(resource) {
  if (resource === "supplements")
    return [
      { key: "name", label: "Producto" },
      { key: "brand", label: "Marca" },
      { key: "price", label: "Precio", type: "money" },
      { key: "stock", label: "Stock" },
      { key: "active", label: "Estado", type: "bool" },
    ];
  if (resource === "orders")
    return [
      { key: "order_number", label: "Número" },
      { key: "customer_name", label: "Cliente" },
      { key: "created_at", label: "Fecha", type: "date" },
      { key: "total", label: "Total", type: "money" },
      { key: "status", label: "Estado" },
    ];
  if (resource === "users")
    return [
      { key: "username", label: "Usuario" },
      { key: "full_name", label: "Nombre" },
      { key: "role", label: "Rol" },
      { key: "active", label: "Estado", type: "bool" },
    ];
  return [
    { key: "name", label: "Nombre" },
    { key: "description", label: "Descripción" },
    { key: "active", label: "Estado", type: "bool" },
  ];
}
function format(value, type) {
  if (type === "money") return money.format(Number(value) || 0);
  if (type === "date") {
    const d = new Date(value);
    return Number.isNaN(d.valueOf())
      ? String(value || "—")
      : dateTime.format(d);
  }
  if (type === "bool") return truthy(value) ? "Activo" : "Inactivo";
  return String(value ?? "—");
}
function openForm(resource, row = {}) {
  form.replaceChildren();
  formStatus.textContent = "";
  formStatus.className = "status";
  document.querySelector("[data-form-title]").textContent =
    `${row.id ? "Editar" : "Crear"} ${labels[resource].toLowerCase()}`;
  const fields =
    resource === "users"
      ? [
          ["username", "Usuario", "text", true],
          ["full_name", "Nombre completo", "text", true],
          ["role", "Rol", "select", true],
          ["password", "Contraseña", "password", !row.id],
          ["active", "Activo", "checkbox", false],
        ]
      : resource === "supplements"
        ? [
            ["name", "Nombre", "text", true],
            ["brand", "Marca", "text", true],
            ["description", "Descripción", "textarea", false],
            ["price", "Precio", "number", true],
            ["stock", "Stock", "number", true],
            ["presentation", "Presentación", "text", false],
            ["flavor", "Sabor", "text", false],
            ["weight", "Contenido/peso", "text", false],
            ["image_url", "URL de imagen", "url", false],
            [
              "category_ids",
              "IDs de categorías (separados por coma)",
              "text",
              false,
            ],
            [
              "goal_ids",
              "IDs de objetivos (separados por coma)",
              "text",
              false,
            ],
            [
              "ingredient_ids",
              "IDs de ingredientes (separados por coma)",
              "text",
              false,
            ],
            ["active", "Activo", "checkbox", false],
          ]
        : [
            ["name", "Nombre", "text", true],
            ["description", "Descripción", "textarea", false],
            ["active", "Activo", "checkbox", false],
          ];
  for (const [f, labelText, type, required] of fields) {
    const label = el("label", {
      text: `${labelText}${required ? " (obligatorio)" : ""}`,
    });
    let input;
    if (type === "select") {
      input = el("select", { attrs: { name: f } });
      input.append(
        el("option", {
          text: "Selecciona un rol",
          attrs: { value: "", disabled: "" },
        }),
      );
      [
        ["ADMINISTRADOR", "Administrador — usuarios y respaldos"],
        ["CAPTURISTA", "Capturista — productos, inventario y pedidos"],
        ["AUDITOR", "Auditor — solo lectura"],
      ].forEach(([value, text]) =>
        input.append(el("option", { text, attrs: { value } })),
      );
    } else if (type === "textarea")
      input = el("textarea", {
        attrs: { name: f, rows: "3", maxlength: "1000" },
      });
    else {
      input = el("input", {
        attrs: {
          name: f,
          type,
          ...(type === "number"
            ? { min: "0", step: f === "price" ? "0.01" : "1" }
            : {}),
        },
      });
    }
    if (required) input.required = true;
    if (type === "checkbox")
      input.checked = row[f] === undefined ? true : truthy(row[f]);
    else input.value = arrayValue(row[f]) || "";
    label.append(input);
    form.append(label);
  }
  const submit = el("button", {
    className: "button primary wide",
    text: "Guardar",
    attrs: { type: "submit" },
  });
  form.append(submit);
  form.onsubmit = (event) => saveForm(event, resource, row.id);
  dialog.showModal();
}
function arrayValue(v) {
  if (Array.isArray(v)) return v.map((x) => x.id ?? x).join(",");
  try {
    const a = JSON.parse(v);
    return Array.isArray(a) ? a.map((x) => x.id ?? x).join(",") : v;
  } catch {
    return v;
  }
}
async function saveForm(event, resource, id) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(form));
  form
    .querySelectorAll("input[type=checkbox]")
    .forEach((i) => (data[i.name] = i.checked));
  for (const key of ["price", "stock"])
    if (key in data) data[key] = Number(data[key]);
  for (const key of ["category_ids", "goal_ids", "ingredient_ids"])
    if (key in data)
      data[key] = String(data[key])
        .split(",")
        .map(Number)
        .filter((n) => Number.isInteger(n) && n > 0);
  const button = form.querySelector("button");
  button.disabled = true;
  try {
    await api(`/api/${resource}${id ? `/${id}` : ""}`, {
      method: id ? "PUT" : "POST",
      body: data,
    });
    dialog.close();
    setStatus("Guardado correctamente.", false, true);
    await show(resource);
  } catch (error) {
    formStatus.textContent = error.message;
    formStatus.className = "status error";
    formStatus.setAttribute("tabindex", "-1");
    formStatus.focus();
  } finally {
    button.disabled = false;
  }
}
async function removeRow(resource, row) {
  if (!confirm(`¿Desactivar ${row.name || row.username}?`)) return;
  try {
    await api(`/api/${resource}/${row.id}`, { method: "DELETE" });
    await show(resource);
  } catch (error) {
    setStatus(error.message, true);
  }
}
function orderStatus(row) {
  form.replaceChildren();
  formStatus.textContent = "";
  formStatus.className = "status";
  const label = el("label", { text: "Estado" }),
    select = el("select", { attrs: { name: "status" } });
  ["PENDIENTE", "CONFIRMADO", "PREPARANDO", "COMPLETADO", "CANCELADO"].forEach(
    (v) => select.append(el("option", { text: v, attrs: { value: v } })),
  );
  select.value = row.status;
  label.append(select);
  form.append(
    label,
    el("button", {
      className: "button primary",
      text: "Actualizar",
      attrs: { type: "submit" },
    }),
  );
  form.onsubmit = async (e) => {
    e.preventDefault();
    try {
      await api(`/api/orders/${row.id}/status`, {
        method: "PUT",
        body: { status: select.value },
      });
      dialog.close();
      await show("orders");
    } catch (error) {
      formStatus.textContent = error.message;
      formStatus.className = "status error";
      formStatus.setAttribute("tabindex", "-1");
      formStatus.focus();
    }
  };
  document.querySelector("[data-form-title]").textContent =
    `Pedido ${row.order_number}`;
  dialog.showModal();
}
function backups() {
  const panel = el("div", { className: "kpi" });
  panel.append(
    el("h2", { text: "Respaldo lógico" }),
    el("p", {
      text: "Exporta esquema y datos de la aplicación a un archivo JSON local con suma SHA-256.",
    }),
  );
  const button = el("button", {
    className: "button primary",
    text: "Generar respaldo",
    attrs: { type: "button" },
  });
  button.addEventListener("click", async () => {
    button.disabled = true;
    setStatus("Generando respaldo paginado…");
    try {
      const result = await api("/api/backups", { method: "POST", body: {} });
      setStatus(
        `Respaldo ${result.file} creado. SHA-256: ${result.sha256}`,
        false,
        true,
      );
    } catch (error) {
      setStatus(error.message, true);
    } finally {
      button.disabled = false;
    }
  });
  panel.append(button);
  content.append(panel);
}
async function logout() {
  try {
    await api("/api/auth/logout", { method: "POST", body: {} });
  } finally {
    setCSRF("");
    location.assign("/login.html");
  }
}
function truthy(v) {
  return v === true || v === 1 || v === "1" || v === "true";
}
function setStatus(message, isError = false, isSuccess = false) {
  status.textContent = message;
  status.className = `status${isError ? " error" : isSuccess ? " success" : ""}`;
}
init();
