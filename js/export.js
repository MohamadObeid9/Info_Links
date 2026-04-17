// ===================== EXPORT =====================
async function exportData() {
  showToast("Exporting…");
  try {
    const [
      programs,
      years,
      semesters,
      courses,
      links,
      extraSections,
      extraLinks,
    ] = await Promise.all([
      sb("programs", "GET", null, null, "*&order=display_order.asc"),
      sb("years", "GET", null, null, "*&order=display_order.asc"),
      sb("semesters", "GET", null, null, "*&order=display_order.asc"),
      sb("courses", "GET", null, null, "*&order=display_order.asc"),
      sb("links", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_sections", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_links", "GET", null, null, "*&order=display_order.asc"),
    ]);

    const payload = {
      exported_at: new Date().toISOString(),
      programs,
      years,
      semesters,
      courses,
      links,
      extra_sections: extraSections,
      extra_links: extraLinks,
    };

    const blob = new Blob([JSON.stringify(payload, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    const date = new Date().toISOString().slice(0, 10);
    a.href = url;
    a.download = `infolinks-backup-${date}.json`;
    a.click();
    URL.revokeObjectURL(url);
    showToast("✅ Backup downloaded!");
  } catch (e) {
    showToast("Export failed: " + e.message, true);
  }
}

// ===================== TOAST =====================
function showToast(msg, isError = false) {
  const t = document.getElementById("toast");
  t.textContent = msg;
  t.className = "toast" + (isError ? " error" : "");
  t.classList.add("show");
  setTimeout(() => t.classList.remove("show"), 3000);
}

// ===================== INIT =====================
document.getElementById("modal").addEventListener("click", (e) => {
  if (e.target === document.getElementById("modal")) closeModal();
});

async function initApp() {
  trackVisit();
  loadAll();
}
// Note: initApp() is called by app.js after all scripts load — do NOT call it here.