// ===================== EXPORT =====================
async function exportData() {
  const btn = document.querySelector(".admin-header .btn-ghost");
  setBtnLoading(btn, true, "Exporting…");
  try {
    const [
      programs,
      years,
      semesters,
      courses,
      links,
      extraSections,
      extraLinks,
      linkClicks,
    ] = await Promise.all([
      sb("programs", "GET", null, null, "*&order=display_order.asc"),
      sb("years", "GET", null, null, "*&order=display_order.asc"),
      sb("semesters", "GET", null, null, "*&order=display_order.asc"),
      sb("courses", "GET", null, null, "*&order=display_order.asc"),
      sb("links", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_sections", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_links", "GET", null, null, "*&order=display_order.asc"),
      sb("link_clicks", "GET", null, null, "*").catch(() => []),
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
      link_clicks: linkClicks,
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
  } finally {
    setBtnLoading(btn, false);
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
  await loadAll();
  // Restore view from URL hash (enables deep-linking)
  const hash = window.location.hash.replace("#", "");
  const validViews = ["home", "report-submit", "feedback", "admin-gate", "admin"];
  if (hash && validViews.includes(hash)) {
    showView(hash);
  }
}
// Note: initApp() is called by app.js after all scripts load — do NOT call it here.