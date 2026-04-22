
// ===================== HELPERS =====================
function esc(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// ===================== THEME =====================
function toggleTheme() {
  AppState.isDark = !AppState.isDark;
  document.documentElement.setAttribute(
    "data-theme",
    AppState.isDark ? "dark" : "light",
  );
  document.getElementById("themeBtn").textContent =
    AppState.isDark ? "🌙" : "☀️";
  // Persist preference
  localStorage.setItem(
    "infolinks_theme",
    AppState.isDark ? "dark" : "light",
  );
}

// ===================== MOBILE =====================
function toggleMobileMenu() {
  document.getElementById("hamburgerBtn").classList.toggle("open");
  document.getElementById("navLinks").classList.toggle("mobile-open");
}
function toggleFilters() {
  document.getElementById("filterToggleBtn").classList.toggle("open");
  document.getElementById("filtersCollapsible").classList.toggle("open");
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
      showView("home");
      search.focus();
      search.select();
    }
  }
  // Escape → close modal
  if (e.key === "Escape") closeModal();
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
 */
function _buildCourseCard(c) {
  const isFav = AppState.favorites.has(String(c.id));
  const linksHtml = c.links.length
    ? c.links
      .map(
        (l) => `
            <a class="link-item"
               data-url="${esc(l.url)}"
               data-link-id="${l.id}"
               onclick="confirmLink(${l.id}); return false;"
               href="#">
              <span class="link-item-main">
                ${getLinkBadge(l.type)}
                <span class="link-label">${esc(l.label)}</span>
                ${l.note ? `<span class="link-note">${esc(l.note)}</span>` : ""}
                <button class="copy-btn" title="Copy link"
                  onclick="event.stopPropagation(); copyLink('${esc(l.url)}')"
                  aria-label="Copy link">⎘</button>
              </span>
              ${getContentTypeChips(l.content_type)}
            </a>`,
      )
      .join("")
    : '<span class="no-links">No links yet — contribute!</span>';

  return `
    <div class="course-card" id="course-card-${c.id}">
      <div class="course-header">
        <div class="course-name">${esc(c.name)}</div>
        <div style="display:flex;align-items:center;gap:6px;">
          ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
          <div class="course-code">${esc(c.code)}</div>
          <button class="fav-btn ${isFav ? "active" : ""}"
            title="${isFav ? "Remove from My Courses" : "Add to My Courses"}"
            onclick="handleFavoriteToggle(${c.id})"
            aria-label="Favorite">★</button>
        </div>
      </div>
      <div class="links-list">${linksHtml}</div>
    </div>`;
}

// ===================== FAVORITES =====================
function handleFavoriteToggle(courseId) {
  toggleFavorite(courseId);
  // Re-render current view to update the star
  if (AppState.currentProg === "favorites") {
    renderCourses(); // will re-render favorites tab
  } else {
    // Just flip the star on the existing card without full re-render
    const btn = document.querySelector(`#course-card-${courseId} .fav-btn`);
    if (btn) {
      const isFav = AppState.favorites.has(String(courseId));
      btn.classList.toggle("active", isFav);
      btn.title = isFav ? "Remove from My Courses" : "Add to My Courses";
    }
  }
}

// ===================== COPY LINK =====================
function copyLink(url) {
  navigator.clipboard.writeText(url).then(() => {
    showToast("Link copied to clipboard! 📋");
  }).catch(() => {
    showToast("Copy failed — try manually.", true);
  });
}

// ===================== BANNER =====================
function toggleImportantNote() {
  document.getElementById("importantNote").classList.toggle("open");
}

const noteTranslations = {
  en: {
    title: "Important Note",
    content: `We are improving the site and need your help! We're adding details about what each link contains. If
                you've tried any link and know its content, click on the <strong>Report / Contribute</strong>
                button. In the <strong>report section</strong>, select the course and the link you want to report.
                In the description, tell us what the link contains (e.g., TD only, Cours only, videos, sessions,
                or if there are missing chapters). Do that for each link you know, and we'll appreciate your help. ❤️`,
  },
  fr: {
    title: "Note Importante",
    content: `Nous améliorons le site et avons besoin de votre aide ! Nous ajoutons des détails sur le contenu de chaque lien. Si vous avez déjà essayé un lien et connaissez son contenu, cliquez sur le bouton <strong>Report / Contribute</strong>. Dans la <strong>section report</strong>, sélectionnez le cours et le lien que vous souhaitez signaler. Dans la description, indiquez ce que contient le lien (ex: TD uniquement, Cours uniquement, vidéos, sessions, ou s'il manque des chapitres). Faites cela pour chaque lien que vous connaissez, et nous apprécierons votre aide. ❤️`,
  },
  ar: {
    title: "ملاحظة هامة",
    content: `نحن نعمل على تحسين الموقع ونحتاج إلى مساعدتك! نضيف تفاصيل حول محتوى كل رابط. إذا قمت بتجربة أي رابط وتعرف محتواه، انقر فوق الزر <strong>Report / Contribute</strong>. في <strong>قسم التقرير (report)</strong>، اختر المادة والرابط الذي تريد الإبلاغ عنه. في الوصف، أخبرنا بما يحتويه الرابط (مثلاً: TD فقط، Cours فقط، فيديو، جلسات، أو إذا كانت هناك فصول مفقودة). افعل ذلك لكل رابط تعرفه، وسنقدر مساعدتك. ❤️`,
  },
};

function setNoteLang(lang) {
  const trans = noteTranslations[lang];
  if (trans) {
    document.getElementById("noteTitleText").textContent = trans.title;
    document.getElementById("noteContentText").innerHTML = trans.content;
    const noteEl = document.getElementById("importantNote");
    if (lang === "ar") {
      noteEl.style.direction = "rtl";
      noteEl.style.textAlign = "right";
    } else {
      noteEl.style.direction = "ltr";
      noteEl.style.textAlign = "left";
    }
  }
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
window.toggleMobileMenu = toggleMobileMenu;
window.toggleFilters = toggleFilters;
window.copyLink = copyLink;
window.setBtnLoading = setBtnLoading;
window.esc = esc;
window.toggleImportantNote = toggleImportantNote;
window.setNoteLang = setNoteLang;
