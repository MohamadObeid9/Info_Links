import { AppState } from "./state.js";
import { trackVisit, apiRequest } from "./supabase.js";
import { setBtnLoading } from "./ui.js";
import { loadAll, onSearch } from "./data.js";
import {
  renderProgTabs,
  renderYearFilters,
  renderSemFilters,
} from "./home.js";

const VALID_VIEWS = ["home", "report-submit", "feedback", "about", "admin-gate", "admin"];

function _getPathView() {
  const path = window.location.pathname.replace("/", "");
  return VALID_VIEWS.includes(path) ? path : "home";
}

// ===================== EXPORT =====================
async function exportData() {
  const btn = document.querySelector(".admin-header .btn-ghost");
  setBtnLoading(btn, true, "Exporting…");
  try {
    const data = await apiRequest("/api/content");
    const payload = {
      exported_at: new Date().toISOString(),
      programs: data.programs || [],
      years: data.years || [],
      semesters: data.semesters || [],
      courses: data.courses || [],
      links: data.links || [],
      extra_sections: data.extra_sections || [],
      extra_links: data.extra_links || [],
    };

    if (!Array.isArray(payload.programs) || payload.programs.length === 0 || !payload.programs[0]?.slug) {
      throw new Error("Content export did not look like the course tree");
    }

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
let toastHideTimer = null;

function showToast(msg, isError = false) {
  const t = document.getElementById("toast");
  t.textContent = msg;
  t.className = "toast" + (isError ? " error" : "");
  t.classList.add("show");
  if (toastHideTimer) clearTimeout(toastHideTimer);
  const text = String(msg);
  const duration = isError
    ? text.includes("(ref:")
      ? 12000
      : 7000
    : 3000;
  toastHideTimer = setTimeout(() => {
    t.classList.remove("show");
    toastHideTimer = null;
  }, duration);
}

// ===================== INIT =====================
document.getElementById("modal").addEventListener("click", (e) => {
  if (e.target === document.getElementById("modal")) window.closeModal?.();
});

function applyHighlightFromURL() {
  const params = new URLSearchParams(window.location.search);
  const raw = (params.get("highlight") || params.get("q") || "").trim();
  if (!raw) return;

  window.showView("home");
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
      const card = document.getElementById(`course-card-${targetId}`);
      card?.classList.add("open");
      card?.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  });
}

async function initApp() {
  // Session and content load in parallel. bootstrapStudentSession records the
  // page view once the student profile (and user id) is known.
  const sessionReady = window.bootstrapStudentSession
    ? window.bootstrapStudentSession()
    : Promise.resolve(trackVisit());

  await loadAll();

  const highlight =
    new URLSearchParams(window.location.search).get("highlight") ||
    new URLSearchParams(window.location.search).get("q");
  if (highlight && highlight.trim()) {
    applyHighlightFromURL();
    await sessionReady;
    return;
  }

  // Restore view from URL path (enables deep-linking with Clean URLs)
  const v = _getPathView();
  if (v !== "home") {
    window.showView(v);
  }
  await sessionReady;
}

window.applyHighlightFromURL = applyHighlightFromURL;
window.initApp = initApp;
window.showToast = showToast;
window.exportData = exportData;

export { showToast, initApp, exportData, applyHighlightFromURL };
