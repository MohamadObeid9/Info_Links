
// ===================== HELPERS =====================
import { AppState, toggleFavorite } from "./state.js";

function esc(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function _isRegisteredStudent() {
  return !!AppState.studentUser && AppState.studentUser.is_guest === false;
}

/** Real href only for signed-in students, so guests cannot read the URL from the browser status bar. */
function _linkHref(url) {
  return _isRegisteredStudent() ? esc(url) : "#";
}

// ===================== THEME =====================
function getSystemDark() {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

function applyTheme(isDark) {
  AppState.isDark = isDark;
  document.documentElement.setAttribute(
    "data-theme",
    isDark ? "dark" : "light",
  );
  const themeBtn = document.getElementById("themeBtn");
  if (themeBtn) themeBtn.textContent = isDark ? "🌙" : "☀️";
  const themeColor = document.querySelector('meta[name="theme-color"]');
  if (themeColor) {
    themeColor.setAttribute("content", isDark ? "#0f0f13" : "#f4f4fb");
  }
}

function applySystemTheme() {
  applyTheme(getSystemDark());
}

function toggleTheme() {
  applyTheme(!AppState.isDark);
}

(function initTheme() {
  localStorage.removeItem("infolinks_theme");
  applySystemTheme();
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener("change", applySystemTheme);
})();

// ===================== MOBILE =====================
function toggleMobileMenu() {
  document.getElementById("hamburgerBtn").classList.toggle("open");
  document.getElementById("navLinks").classList.toggle("mobile-open");
}
function toggleFilters() {
  document.getElementById("filterToggleBtn").classList.toggle("open");
  document.getElementById("filtersCollapsible").classList.toggle("open");
}

const MOBILE_MQ = "(max-width: 768px)";

function isMobileView() {
  return window.matchMedia(MOBILE_MQ).matches;
}

function adminTd(label, inner, extraAttrs = "") {
  return `<td data-label="${esc(label)}"${extraAttrs}>${inner}</td>`;
}

function adminCell(role, label, inner) {
  return `<td class="${role}" data-label="${esc(label)}">${inner}</td>`;
}

document.addEventListener("click", (e) => {
  if (e.target.classList.contains("nav-btn")) {
    document.getElementById("hamburgerBtn").classList.remove("open");
    document.getElementById("navLinks").classList.remove("mobile-open");
  }
});

// ===================== KEYBOARD SHORTCUT =====================
document.addEventListener("keydown", (e) => {
  // "/" or Ctrl+K → focus search (only on home view)
  const tag = document.activeElement.tagName;
  const isTyping =
    tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
  if (isTyping) return;
  if (e.key === "/" || (e.ctrlKey && e.key === "k")) {
    e.preventDefault();
    const search = document.getElementById("searchInput");
    if (search) {
      window.showView("home");
      search.focus();
      search.select();
    }
  }
  // Escape → close modal
  if (e.key === "Escape") window.closeModal?.();
});

// Helper: find all courses that share the same code as the given courseId
function _findSharedCourses(courseId) {
  let targetCode = null;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === courseId) targetCode = c.code;
        }),
      ),
    ),
  );
  if (!targetCode) return { code: null, siblings: [] };
  const siblings = [];
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id !== courseId && c.code === targetCode)
            siblings.push({
              ...c,
              prog: p.name,
              year: y.name,
              sem: s.name,
            });
        }),
      ),
    ),
  );
  return { code: targetCode, siblings };
}

// ===================== BADGE =====================
function getLinkBadge(type) {
  if (type === "telegram") return '<span class="link-badge badge-tg">TG</span>';
  if (type === "drive") return '<span class="link-badge badge-drive">GD</span>';
  if (type === "classroom")
    return '<span class="link-badge badge-classroom">GC</span>';
  return '<span class="link-badge badge-other">OT</span>';
}

// ===================== CONTENT TYPE CHIP =====================
const CONTENT_TYPE_META = {
  td: { label: "TD", emoji: "✏️", cls: "ct-td" },
  cours: { label: "Cours", emoji: "📄", cls: "ct-cours" },
  videos: { label: "Videos", emoji: "🎬", cls: "ct-videos" },
  sessions: { label: "Sessions", emoji: "🎤", cls: "ct-sessions" },
  exams: { label: "Exams", emoji: "📝", cls: "ct-exams" },
  other: { label: "Other", emoji: "📦", cls: "ct-other" },
};

/**
 * Render content-type chips. Accepts a comma-separated string (e.g. "td,cours")
 * or a single value. Returns wrapped HTML for consistent layout.
 */
function getContentTypeChips(contentType) {
  if (!contentType) return "";
  const types = String(contentType).split(",").map(t => t.trim()).filter(Boolean);
  if (!types.length) return "";
  const chips = types.map(t => {
    const meta = CONTENT_TYPE_META[t];
    if (!meta) return "";
    return `<span class="content-chip ${meta.cls}" title="${meta.label}">${meta.emoji} ${meta.label}</span>`;
  }).filter(Boolean).join("");
  if (!chips) return "";
  return `<span class="content-chips-wrap">${chips}</span>`;
}
// Backward compat alias
function getContentTypeChip(ct) { return getContentTypeChips(ct); }

// ===================== COURSE CARD BUILDER =====================
/**
 * Builds the HTML string for a single course card.
 * Used by both renderCourses (filtered) and renderAllCourses (all).
 * opts.path is shown on mobile search results (program · year · semester).
 */
function _buildCourseCard(c, opts = {}) {
  const isFav = AppState.favorites.has(String(c.id));
  const path = opts.path || "";
  const linksHtml = c.links.length
    ? c.links
      .map(
        (l) => `
            <a class="link-item"
               data-url="${esc(l.url)}"
               data-link-id="${l.id}"
               data-link-kind="link"
               href="${_linkHref(l.url)}">
              <span class="link-item-main">
                ${getLinkBadge(l.type)}
                <span class="link-label">${esc(l.label)}</span>
                ${l.note ? `<span class="link-note">${esc(l.note)}</span>` : ""}
                <button class="copy-btn" title="Copy link"
                  aria-label="Copy link">⎘</button>
              </span>
              ${getContentTypeChips(l.content_type)}
            </a>`,
      )
      .join("")
    : '<span class="no-links">No links yet — contribute!</span>';

  return `
    <div class="course-card" id="course-card-${c.id}">
      <div class="course-header" data-toggle-course="${c.id}">
        <h2 class="course-name">${esc(c.name)}</h2>
        <div class="course-header-side">
          <div class="course-header-tags">
            ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
            <h3 class="course-code">${esc(c.code)}</h3>
            ${path ? `<span class="course-path">${esc(path)}</span>` : ""}
          </div>
          <button class="fav-btn ${isFav ? "active" : ""}"
            title="${isFav ? "Remove from My Courses" : "Add to My Courses"}"
            onclick="handleFavoriteToggle(${c.id})"
            aria-label="Favorite">★</button>
          <span class="course-chev" aria-hidden="true">›</span>
        </div>
      </div>
      <div class="links-list">${linksHtml}</div>
    </div>`;
}

// ===================== FAVORITES =====================
function _paintFavorite(courseId) {
  if (AppState.currentProg === "favorites") {
    window.renderCourses();
    return;
  }
  const btn = document.querySelector(`#course-card-${courseId} .fav-btn`);
  if (btn) {
    const isFav = AppState.favorites.has(String(courseId));
    btn.classList.toggle("active", isFav);
    btn.title = isFav ? "Remove from My Courses" : "Add to My Courses";
  }
}

/** Optimistic star toggle; the server is the source of truth so failures roll back. */
async function handleFavoriteToggle(courseId) {
  const retry = () => handleFavoriteToggle(courseId);
  if (!window.requireStudent(retry)) return;

  const added = !AppState.favorites.has(String(courseId));
  toggleFavorite(courseId);
  _paintFavorite(courseId);

  try {
    await window.syncFavorite(courseId, added);
  } catch (err) {
    toggleFavorite(courseId);
    _paintFavorite(courseId);
    if (window.handleStudentAuthError?.(err, retry)) return;
    window.logApiError?.(err, "syncFavorite");
    window.showToast(
      window.formatApiError?.(err, "Could not save your favorites.") ||
      "Could not save your favorites.",
      true,
    );
  }
}

// ===================== COPY LINK =====================
function copyLink(url) {
  navigator.clipboard.writeText(url).then(() => {
    window.showToast("Link copied to clipboard! 📋");
  }).catch(() => {
    window.showToast("Copy failed — try manually.", true);
  });
}

// ===================== LOADING STATES =====================
/**
 * Toggle a button's loading state.
 * @param {HTMLButtonElement} btn
 * @param {boolean} loading
 * @param {string} [loadingText]
 */
function setBtnLoading(btn, loading, loadingText = "…") {
  if (!btn) return;
  if (loading) {
    btn.dataset.origText = btn.textContent;
    btn.textContent = loadingText;
    btn.disabled = true;
    btn.classList.add("btn-loading");
  } else {
    btn.textContent = btn.dataset.origText || btn.textContent;
    btn.disabled = false;
    btn.classList.remove("btn-loading");
  }
}

window.toggleTheme = toggleTheme;
window.applyTheme = applyTheme;
window.applySystemTheme = applySystemTheme;
window.toggleMobileMenu = toggleMobileMenu;
window.toggleFilters = toggleFilters;
window.copyLink = copyLink;
window.setBtnLoading = setBtnLoading;
window.esc = esc;
window.handleFavoriteToggle = handleFavoriteToggle;
window.isMobileView = isMobileView;

export {
  esc,
  _buildCourseCard,
  _findSharedCourses,
  getLinkBadge,
  getContentTypeChips,
  getContentTypeChip,
  setBtnLoading,
  toggleTheme,
  applyTheme,
  applySystemTheme,
  toggleMobileMenu,
  toggleFilters,
  copyLink,
  handleFavoriteToggle,
  _linkHref,
  isMobileView,
  adminTd,
  adminCell,
};
