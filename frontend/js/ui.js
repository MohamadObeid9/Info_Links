
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
/** One entry per favorite course id, even when the course is offered in several programs. */
function collectFavoriteCourses(query = "") {
  const q = query.toLowerCase().trim();
  const favIds = AppState.favorites;
  const byId = new Map();
  AppState.dbPrograms.forEach((prog) => {
    prog.years.forEach((year) => {
      year.sems.forEach((sem) => {
        (sem.courses || []).forEach((c) => {
          if (!favIds.has(String(c.id))) return;
          if (
            q &&
            !c.name.toLowerCase().includes(q) &&
            !c.code.toLowerCase().includes(q)
          ) {
            return;
          }
          const path = `${prog.name} · ${year.name} · ${sem.name}`;
          const existing = byId.get(c.id);
          if (existing) {
            if (!existing.paths.includes(path)) existing.paths.push(path);
            return;
          }
          byId.set(c.id, { course: c, paths: [path] });
        });
      });
    });
  });
  return [...byId.values()];
}

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

const TELEGRAM_PROMOTE_URL = "https://t.me/Info_Links_Services_Guide";
const TELEGRAM_CONTRIBUTE_URL = "https://t.me/Info_Links_Contributing_Guide";

function hintLink(url, label) {
  return `<a class="hint-link" href="${_linkHref(url)}" data-url="${esc(url)}">${esc(label)}</a>`;
}

const FAVORITES_HINT =
  "Mark the courses you use most with ★ to reach them more easily in the My Courses section.";

const COMMUNITY_PROMOTE_HINT = `Want to promote your service on Info Links? Contact us on ${hintLink(TELEGRAM_PROMOTE_URL, "Telegram")} to know more.`;

const CONTRIBUTE_HINT = `Want to help and contribute to Info Links? Visit our ${hintLink(TELEGRAM_CONTRIBUTE_URL, "Telegram guide")} to know more on how you can help us make Info Links better.`;

function linkTypesLegendHtml() {
  return `
    <div class="hint-link-types">
      <span class="hint-legend-item">${getLinkBadge("telegram")} Telegram</span>
      <span class="hint-legend-item">${getLinkBadge("drive")} Google Drive</span>
      <span class="hint-legend-item">${getLinkBadge("classroom")} Google Classroom</span>
      <span class="hint-legend-item">${getLinkBadge("other")} Other Type</span>
      <span class="hint-legend-item"><span class="optional-tag">OPTIONAL</span> Optional course</span>
    </div>`;
}

function hintCardHtml(title, bodyHtml, extraClass = "") {
  const classes = ["fav-hint", extraClass].filter(Boolean).join(" ");
  return `<div class="${classes}"><div class="fav-hint-title">${esc(title)}</div><div class="fav-hint-body">${bodyHtml}</div></div>`;
}

const FAVORITES_HINT_CARD = hintCardHtml("Favorites", FAVORITES_HINT);
const COMMUNITY_HINT_CARD = hintCardHtml("Community Services", COMMUNITY_PROMOTE_HINT);
const CONTRIBUTE_HINT_CARD = hintCardHtml("Contributing", CONTRIBUTE_HINT);
const LINK_TYPES_HINT_CARD = hintCardHtml("Link types", linkTypesLegendHtml(), "fav-hint--link-types");

function tipsSectionHtml() {
  return `
    <div class="tips-section">
      ${FAVORITES_HINT_CARD}
      ${COMMUNITY_HINT_CARD}
      ${CONTRIBUTE_HINT_CARD}
      ${LINK_TYPES_HINT_CARD}
    </div>`;
}

function homeSectionHeading(title) {
  return `<h2 class="home-section-title">${title}</h2>`;
}

function sectionInlineHintHtml(bodyHtml) {
  return `<div class="section-inline-hint">${bodyHtml}</div>`;
}

function setSectionHint(html) {
  const el = document.getElementById("sectionHint");
  if (!el) return;
  if (!html) {
    el.hidden = true;
    el.replaceChildren();
    return;
  }
  el.innerHTML = html;
  el.hidden = false;
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
  FAVORITES_HINT,
  COMMUNITY_PROMOTE_HINT,
  FAVORITES_HINT_CARD,
  COMMUNITY_HINT_CARD,
  collectFavoriteCourses,
  setSectionHint,
  tipsSectionHtml,
  homeSectionHeading,
  sectionInlineHintHtml,
};
