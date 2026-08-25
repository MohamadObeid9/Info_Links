import { AppState } from "./state.js";
import { _loadCache, _saveCache } from "./cache.js";
import { apiRequest, formatApiError, sb } from "./supabase.js";
import { showSkeleton } from "./skeleton.js";
import { esc, isMobileView } from "./ui.js";
import {
  renderProgTabs,
  renderYearFilters,
  renderSemFilters,
  renderCourses,
  renderExtra,
} from "./home.js";
import { initMobileHomeState } from "./mobile-home.js";

// ===================== LOAD DATA =====================
function _buildTree(
  programs,
  years,
  semesters,
  courses,
  links,
  extraSections,
  extraLinks,
) {
  AppState.dbPrograms = programs.map((p) => ({
    ...p,
    years: years
      .filter((y) => y.program_id === p.id)
      .map((y) => ({
        ...y,
        sems: semesters
          .filter((s) => s.year_id === y.id)
          .map((s) => ({
            ...s,
            courses: courses
              .filter((c) => c.semester_id === s.id)
              .map((c) => ({
                ...c,
                links: links.filter((l) => l.course_id === c.id)
                  .sort(_naturalLinkSort),
              })),
          })),
      })),
  }));

  AppState.dbExtra = extraSections.map((sec) => ({
    ...sec,
    links: extraLinks.filter((l) => l.section_id === sec.id)
      .sort(_naturalLinkSort),
  }));

  AppState.courseById = new Map();
  AppState.linkById = new Map();
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          AppState.courseById.set(c.id, c);
          c.links.forEach((l) => AppState.linkById.set(l.id, l));
        }),
      ),
    ),
  );
  AppState.dbExtra.forEach((sec) => sec.links.forEach((l) => AppState.linkById.set(l.id, l)));
}

// Natural sort: extract trailing number from label for numeric ordering
function _naturalLinkSort(a, b) {
  const numA = parseInt((a.label || "").match(/(\d+)\s*$/)?.[1]) || 0;
  const numB = parseInt((b.label || "").match(/(\d+)\s*$/)?.[1]) || 0;
  if (numA !== numB) return numA - numB;
  return (a.display_order || 0) - (b.display_order || 0);
}

function _contentPayload(data) {
  return {
    programs: data.programs || [],
    years: data.years || [],
    semesters: data.semesters || [],
    courses: data.courses || [],
    links: data.links || [],
    extra_sections: data.extra_sections || [],
    extra_links: data.extra_links || [],
  };
}

function _applyContent(payload) {
  _buildTree(
    payload.programs,
    payload.years,
    payload.semesters,
    payload.courses,
    payload.links,
    payload.extra_sections,
    payload.extra_links,
  );
}

async function _fetchAndCacheContent() {
  const payload = _contentPayload(await apiRequest("/api/content"));
  if (!AppState.adminLoggedIn) _saveCache(payload);
  return payload;
}

function _renderAfterLoad({ resetMobile = true } = {}) {
  if (resetMobile) initMobileHomeState();
  if (!AppState.currentProg) AppState.currentProg = "all";
  document.getElementById("extraSection").style.display = "none";
  if (isMobileView()) {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("coursesOutput").style.display = "";
    renderProgTabs();
    renderCourses();
    _populateCourseDatalist();
    return;
  }
  if (AppState.currentProg === "all") {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("extraSection").style.display = "";
  }
  renderProgTabs();
  renderYearFilters();
  renderSemFilters();
  renderCourses();
  if (AppState.currentProg === "all") renderExtra();
  // Populate the course datalist for Report/Contribute autocomplete
  _populateCourseDatalist();
}

function _populateCourseDatalist() {
  const dl = document.getElementById("courseDatalist");
  if (!dl) return;
  const names = new Set();
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          names.add(`${c.name} (${c.code})`);
        }),
      ),
    ),
  );
  dl.innerHTML = [...names]
    .sort()
    .map((n) => `<option value="${esc(n)}">`)
    .join("");
}

async function loadAll() {
  const output = document.getElementById("coursesOutput");
  const extra = document.getElementById("extraSection");
  const cached = AppState.adminLoggedIn ? null : _loadCache();

  if (cached) {
    _applyContent(cached.data);
    output.dataset.loaded = "1";
    _renderAfterLoad();
    if (!cached.stale) return;
    try {
      const fresh = await _fetchAndCacheContent();
      if (JSON.stringify(cached.data) !== JSON.stringify(fresh)) {
        _applyContent(fresh);
        _renderAfterLoad({ resetMobile: false });
      }
    } catch {
      // Keep showing the stale payload if the refresh fails.
    }
    return;
  }

  const isFirstLoad = !output.dataset.loaded;
  if (isFirstLoad) {
    showSkeleton();
  } else {
    output.innerHTML =
      '<div class="loader"><div class="spinner"></div> Loading…</div>';
  }
  extra.innerHTML = "";
  try {
    const payload = await _fetchAndCacheContent();
    _applyContent(payload);
    output.dataset.loaded = "1";
    _renderAfterLoad();
  } catch (e) {
    const message = formatApiError(e, "Failed to fetch from backend");
    output.innerHTML =
      `<div class="empty">⚠️ Failed to load data: ${esc(message)}</div>`;
  }
}

async function loadReportsBadges() {
  try {
    const [reports, contribs, feedback] = await Promise.all([
      sb("reports", "GET", null, null, "id,status"),
      sb("contributions", "GET", null, null, "id,status"),
      sb("feedback", "GET", null, null, "id,status"),
    ]);
    const openR = reports.filter((r) => r.status === "open").length;
    const pendC = contribs.filter((c) => c.status === "pending").length;
    const newF = feedback.filter((f) => f.status === "new").length;
    const rb = document.getElementById("reportBadge"),
      cb = document.getElementById("contribBadge"),
      fb = document.getElementById("feedbackBadge");
    rb.style.display = openR ? "inline" : "none";
    rb.textContent = openR;
    rb.classList.toggle("is-alert", openR > 0);
    cb.style.display = pendC ? "inline" : "none";
    cb.textContent = pendC;
    cb.classList.toggle("is-alert", pendC > 0);
    fb.style.display = newF ? "inline" : "none";
    fb.textContent = newF;
    fb.classList.toggle("is-alert", newF > 0);
    _paintAdminInboxHints(openR, pendC, newF);
  } catch (e) {}
}

function _paintAdminInboxHints(openR, pendC, newF) {
  const select = document.getElementById("adminTabSelect");
  if (select) {
    const labels = {
      feedback: newF ? `Feedback (${newF})` : "Feedback",
      reports: openR ? `Reports (${openR})` : "Reports",
      contributions: pendC ? `Contributions (${pendC})` : "Contributions",
    };
    [...select.options].forEach((opt) => {
      if (labels[opt.value]) opt.textContent = labels[opt.value];
    });
  }

  const el = document.getElementById("adminMobileAlerts");
  if (!el) return;
  const items = [
    newF > 0 && { tab: "feedback", label: "Feedback", n: newF },
    openR > 0 && { tab: "reports", label: "Reports", n: openR },
    pendC > 0 && { tab: "contributions", label: "Contributions", n: pendC },
  ].filter(Boolean);
  el.hidden = items.length === 0;
  el.innerHTML = items
    .map(
      (i) =>
        `<button type="button" class="admin-alert-chip" data-admin-tab="${i.tab}">${i.label} <span class="badge is-alert">${i.n}</span></button>`,
    )
    .join("");
}

function onSearch() {
  window.trackSearch?.(document.getElementById("searchInput")?.value || "");
  if (isMobileView()) {
    window.renderCourses();
    return;
  }
  if (AppState.currentProg === "extra") window.renderExtra();
  else if (AppState.currentProg === "all") {
    window.renderCourses();
    window.renderExtra();
  } else window.renderCourses();
}
window.loadAll = loadAll;
window.onSearch = onSearch;
window.loadReportsBadges = loadReportsBadges;

export { loadAll, onSearch, loadReportsBadges };
