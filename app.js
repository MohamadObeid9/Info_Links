async function loadHtml(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to load HTML: ${url}`);
  return res.text();
}

function loadScript(url) {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = url;
    script.async = false; // Ensures scripts execute in the order they are added
    script.onload = resolve;
    script.onerror = () => reject(new Error(`Failed to load script: ${url}`));
    document.body.appendChild(script);
  });
}

async function bootstrap() {
  const appRoot = document.getElementById("app");
  if (!appRoot) throw new Error("Missing #app container");

  appRoot.innerHTML = await loadHtml("html/body.html");

  const scripts = [
    "js/skeleton.js",
    "js/supabase.js",
    "js/state.js",
    "js/ui.js",
    "js/cache.js",
    "js/data.js",
    "js/home.js",
    "js/views.js",
    "js/admin.js",
    "js/feedback.js",
    "js/modals.js",
    "js/export.js",
  ];

  await Promise.all(scripts.map(loadScript));

  if (typeof initApp === "function") {
    initApp().catch((err) => console.error("App init failed:", err));
  }
}

window.addEventListener("DOMContentLoaded", () => {
  bootstrap().catch((err) => {
    console.error(err);
    const appRoot = document.getElementById("app");
    if (appRoot)
      appRoot.innerHTML = `<div class="empty">⚠️ ${err.message}</div>`;
  });
});
