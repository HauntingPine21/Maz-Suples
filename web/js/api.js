function cookie(name) {
  return (
    document.cookie
      .split("; ")
      .find((v) => v.startsWith(`${name}=`))
      ?.split("=")
      .slice(1)
      .join("=") || ""
  );
}
let csrfToken =
  sessionStorage.getItem("maz_csrf") || decodeURIComponent(cookie("maz_csrf"));

export function setCSRF(token) {
  csrfToken = token || "";
  if (token) sessionStorage.setItem("maz_csrf", token);
  else sessionStorage.removeItem("maz_csrf");
}
export async function api(path, { method = "GET", body, signal } = {}) {
  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (!csrfToken) csrfToken = decodeURIComponent(cookie("maz_csrf"));
  if (!["GET", "HEAD", "OPTIONS"].includes(method) && csrfToken)
    headers["X-CSRF-Token"] = csrfToken;
  const response = await fetch(path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: "same-origin",
    signal,
  });
  if (response.status === 204) return null;
  const payload = await response
    .json()
    .catch(() => ({ error: { message: "Respuesta inválida del servidor" } }));
  if (!response.ok) {
    const error = new Error(
      payload.error?.message || "No fue posible completar la operación",
    );
    error.status = response.status;
    error.code = payload.error?.code;
    throw error;
  }
  return payload;
}
export async function download(path, { method = "POST", body = {} } = {}) {
  const headers = {
    Accept: "application/sql",
    "Content-Type": "application/json",
  };
  if (!csrfToken) csrfToken = decodeURIComponent(cookie("maz_csrf"));
  if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
  const response = await fetch(path, {
    method,
    headers,
    body: JSON.stringify(body),
    credentials: "same-origin",
  });
  if (!response.ok) {
    const payload = await response
      .json()
      .catch(() => ({ error: { message: "No fue posible descargar el archivo" } }));
    const error = new Error(
      payload.error?.message || "No fue posible descargar el archivo",
    );
    error.status = response.status;
    error.code = payload.error?.code;
    throw error;
  }
  const disposition = response.headers.get("Content-Disposition") || "";
  const match = disposition.match(/filename="?([^";]+)"?/i);
  return {
    blob: await response.blob(),
    filename: (match?.[1] || "maz-suplementos.sql").replace(
      /[^a-zA-Z0-9._-]/g,
      "_",
    ),
  };
}
export const money = new Intl.NumberFormat("es-MX", {
  style: "currency",
  currency: "MXN",
});
export const dateTime = new Intl.DateTimeFormat("es-MX", {
  dateStyle: "medium",
  timeStyle: "short",
});
export function el(tag, { className, text, attrs } = {}) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  if (attrs)
    for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, v);
  return node;
}
export function safeImage(value) {
  try {
    const u = new URL(value);
    return ["http:", "https:"].includes(u.protocol)
      ? u.href
      : "/assets/product-placeholder.svg";
  } catch {
    return "/assets/product-placeholder.svg";
  }
}
