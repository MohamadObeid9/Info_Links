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

function applyHighlightFromURL() {
  const params = new URLSearchParams(window.location.search);
  const raw = (params.get("highlight") || params.get("q") || "").trim();
  if (!raw) return;

  showView("home");
  AppState.currentProg = "all";
  document.querySelector(".filter-row").style.display = "none";
  const extra = document.getElementById("extraSection");
  if (extra) extra.style.display = "";
  renderProgTabs();
  renderYearFilters();
  renderSemFilters();

  const search = document.getElementById("searchInput");
  if (search) search.value = raw;
  onSearch();

  requestAnimationFrame(() => {
    const q = raw.toLowerCase();
    let targetId = null;
    AppState.courseById.forEach((c) => {
      if (targetId) return;
      if (
        c.code.toLowerCase() === q ||
        c.code.toLowerCase().includes(q) ||
        c.name.toLowerCase().includes(q)
      ) {
        targetId = c.id;
      }
    });
    if (targetId) {
      document
        .getElementById(`course-card-${targetId}`)
        ?.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  });
}

async function initApp() {
  trackVisit();
  await loadAll();

  const highlight =
    new URLSearchParams(window.location.search).get("highlight") ||
    new URLSearchParams(window.location.search).get("q");
  if (highlight && highlight.trim()) {
    applyHighlightFromURL();
    return;
  }

  // Restore view from URL path (enables deep-linking with Clean URLs)
  const v = _getPathView();
  if (v !== "home") {
    showView(v);
  }
}
window.applyHighlightFromURL = applyHighlightFromURL;
window.initApp = initApp;
window.showToast = showToast;
window.exportData = exportData;
