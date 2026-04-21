// ===================== STATE =====================
// All mutable application state lives here as a single object.
// Access via AppState.xxx; mutate directly: AppState.xxx = yyy.

const AppState = {
  sbToken: null,
  adminLoggedIn: false,
  currentAdminTab: "courses",
  adminSearch: "",
  adminFilterProg: "all",
  adminFilterYear: "all",
  adminFilterSem: "all",
  _pendingCourseEdit: null,
  _pendingLinkOp: null,
  isDark: false,
  currentProg: null,
  currentYear: "all",
  currentSem: "all",
  dbPrograms: [],
  analyticsRange: "30",
  dbExtra: [],
  courseById: new Map(),
  linkById: new Map(),
  favorites: new Set(), // Set of course IDs (strings from localStorage)
};

// ── Theme: read from localStorage first, then system preference ──────────────
(function initTheme() {
  const saved = localStorage.getItem("infolinks_theme");
  if (saved) {
    AppState.isDark = saved === "dark";
  } else {
    AppState.isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  }
  document.documentElement.setAttribute(
    "data-theme",
    AppState.isDark ? "dark" : "light",
  );
  const themeBtn = document.getElementById("themeBtn");
  if (themeBtn) themeBtn.textContent = AppState.isDark ? "🌙" : "☀️";
})();

// ── Favorites: restore from localStorage ─────────────────────────────────────
(function initFavorites() {
  try {
    const raw = localStorage.getItem("infolinks_favorites");
    if (raw) {
      const arr = JSON.parse(raw);
      AppState.favorites = new Set(arr.map(String));
    }
  } catch (e) {
    AppState.favorites = new Set();
  }
})();

function saveFavorites() {
  try {
    localStorage.setItem(
      "infolinks_favorites",
      JSON.stringify([...AppState.favorites]),
    );
  } catch (e) { }
}

function toggleFavorite(courseId) {
  const key = String(courseId);
  if (AppState.favorites.has(key)) {
    AppState.favorites.delete(key);
  } else {
    AppState.favorites.add(key);
  }
  saveFavorites();
}

// ── Backward-compat shims so legacy code reading bare vars still works ────────
// (These will be cleaned up gradually; new code should use AppState.xxx)
Object.defineProperties(window, {
  sbToken: { get: () => AppState.sbToken, set: v => { AppState.sbToken = v; } },
  adminLoggedIn: { get: () => AppState.adminLoggedIn, set: v => { AppState.adminLoggedIn = v; } },
  currentAdminTab: { get: () => AppState.currentAdminTab, set: v => { AppState.currentAdminTab = v; } },
  adminSearch: { get: () => AppState.adminSearch, set: v => { AppState.adminSearch = v; } },
  adminFilterProg: { get: () => AppState.adminFilterProg, set: v => { AppState.adminFilterProg = v; } },
  adminFilterYear: { get: () => AppState.adminFilterYear, set: v => { AppState.adminFilterYear = v; } },
  adminFilterSem: { get: () => AppState.adminFilterSem, set: v => { AppState.adminFilterSem = v; } },
  _pendingCourseEdit: { get: () => AppState._pendingCourseEdit, set: v => { AppState._pendingCourseEdit = v; } },
  _pendingLinkOp: { get: () => AppState._pendingLinkOp, set: v => { AppState._pendingLinkOp = v; } },
  isDark: { get: () => AppState.isDark, set: v => { AppState.isDark = v; } },
  currentProg: { get: () => AppState.currentProg, set: v => { AppState.currentProg = v; } },
  currentYear: { get: () => AppState.currentYear, set: v => { AppState.currentYear = v; } },
  currentSem: { get: () => AppState.currentSem, set: v => { AppState.currentSem = v; } },
  dbPrograms: { get: () => AppState.dbPrograms, set: v => { AppState.dbPrograms = v; } },
  analyticsRange: { get: () => AppState.analyticsRange, set: v => { AppState.analyticsRange = v; } },
  dbExtra: { get: () => AppState.dbExtra, set: v => { AppState.dbExtra = v; } },
});

window.AppState = AppState;