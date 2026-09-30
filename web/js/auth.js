import { api, setCSRF } from "./api.js";
const form = document.querySelector("[data-login-form]");
const status = document.querySelector("[data-login-status]");
form?.addEventListener("submit", async (event) => {
  event.preventDefault();
  status.className = "status";
  status.textContent = "Comprobando credenciales…";
  const submit = form.querySelector("button");
  submit.disabled = true;
  try {
    const data = await api("/api/auth/login", {
      method: "POST",
      body: Object.fromEntries(new FormData(form)),
    });
    setCSRF(data.csrf_token);
    status.className = "status success";
    status.textContent = "Acceso correcto. Abriendo el panel…";
    location.assign("/admin/");
  } catch (error) {
    status.className = "status error";
    status.textContent = error.message;
  } finally {
    submit.disabled = false;
  }
});
