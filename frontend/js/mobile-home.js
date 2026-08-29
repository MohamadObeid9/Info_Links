import { AppState } from "./state.js";
import { esc, isMobileView, _buildCourseCard, getLinkBadge, getContentTypeChips, _linkHref } from "./ui.js";

const MOBILE_MQ = "(max-width: 768px)";

function coerceId(raw) {
  if (raw === "extra" || raw === "favorites" || raw === "all") return raw;
  const n = Number(raw);
  return Number.isFinite(n) && String(n) === String(raw) ? n : raw;
}

function idsEqual(a, b) {
  return String(a) === String(b);
}

function isRealProgram(id) {
  return id != null && id !== "all" && id !== "extra" && id !== "favorites";
}

function searchQuery() {
  return document.getElementById("searchInput")?.value.toLowerCase().trim() || "";
}

function findProgram(id) {
  return AppState.dbPrograms.find((p) => idsEqual(p.id, id));
}

function hideExtra() {
  const extra = document.getElementById("extraSection");
  const courses = document.getElementById("coursesOutput");
  if (extra) extra.style.display = "none";
  if (courses) courses.style.display = "";
}

function showExtraOnly() {
  const extra = document.getElementById("extraSection");
  const courses = document.getElementById("coursesOutput");
  if (courses) courses.style.display = "none";
  if (extra) extra.style.display = "";
}

function chipsHtml(labels, backStep) {
  return `
    <div class="mobile-chips">
      ${labels.map((l) => `<span class="mobile-chip">${esc(l)}</span>`).join("")}
      <button type="button" class="mobile-chip-change" data-mobile-back="${backStep}">Change</button>
    </div>`;
}

function extraMatches(q) {
  if (!q) return [];
  return AppState.dbExtra.filter(
    (r) =>
      r.title.toLowerCase().includes(q) ||
      r.links.some((l) => l.label.toLowerCase().includes(q)),
  );
}

function extraCardsHtml(sections) {
  return sections
    .map(
      (r) => `
        <div class="extra-section">
          <div class="extra-title"><span>${esc(r.icon)}</span>${esc(r.title)}</div>
          <div class="links-list">
            ${r.links.length
          ? r.links
            .map(
              (l) => `
              <a class="link-item"
                 data-url="${esc(l.url)}"
                 data-link-id="${l.id}"
                 data-link-kind="extra_link"
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
          : '<span class="no-links">No links yet — contribute!</span>'}
          </div>
        </div>`,
    )
    .join("");
}

function collectSearchHits(q) {
  const hits = [];
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (
            c.name.toLowerCase().includes(q) ||
            c.code.toLowerCase().includes(q)
          ) {
            hits.push({ course: c, path: `${p.name} · ${y.name} · ${s.name}` });
          }
        }),
      ),
    ),
  );
  return hits;
}

function renderMobileSearch(q) {
  hideExtra();
  const hits = collectSearchHits(q);
  const extras = extraMatches(q);
  let html = `<div class="mobile-section-label">${hits.length} course${hits.length === 1 ? "" : "s"}</div>`;
  html += hits.length
    ? `<div class="courses-grid">${hits.map((h) => _buildCourseCard(h.course, { path: h.path })).join("")}</div>`
    : '<div class="empty">No course matches that. Try a code like NFA035.</div>';
  if (extras.length) {
    html += `<div class="mobile-section-label">Extra resources</div>${extraCardsHtml(extras)}`;
  }
  document.getElementById("coursesOutput").innerHTML = html;
}

function renderMobileProgramPicker() {
  hideExtra();
  const programs = AppState.dbPrograms
    .map(
      (p) => `
        <button type="button" class="pick-card" data-mobile-prog="${p.id}">
          <span>
            <span class="pick-title">${esc(p.name)}</span>
            <small>${p.years.length} year${p.years.length === 1 ? "" : "s"}</small>
          </span>
          <span class="pick-chev">›</span>
        </button>`,
    )
    .join("");

  document.getElementById("coursesOutput").innerHTML = `
    <p class="mobile-hint">Search if you know the code — or pick your program to browse this semester.</p>
    <div class="mobile-section-label">Your program</div>
    ${programs}
    <button type="button" class="pick-card" data-mobile-prog="extra">
      <span>
        <span class="pick-title">📦 Extra resources</span>
        <small>Outside the course tree</small>
      </span>
      <span class="pick-chev">›</span>
    </button>
    <button type="button" class="pick-card" data-mobile-prog="community">
      <span>
        <span class="pick-title">🤝 Community Services</span>
        <small>Small student businesses & services</small>
      </span>
      <span class="pick-chev">›</span>
    </button>
    <button type="button" class="pick-card" data-mobile-prog="favorites">
      <span>
        <span class="pick-title">⭐ My Courses</span>
        <small>Saved on this account</small>
      </span>
      <span class="pick-chev">›</span>
    </button>`;
}

function renderMobileYearPicker() {
  hideExtra();
  const prog = findProgram(AppState.currentProg);
  if (!prog) {
    AppState.mobileStep = "program";
    renderMobileProgramPicker();
    return;
  }

  const years = prog.years
    .map((y) => {
      const sems = (y.sems || [])
        .map(
          (s) =>
            `<button type="button" class="mobile-sem-btn" data-mobile-year="${y.id}" data-mobile-sem="${s.id}">${esc(s.name)}</button>`,
        )
        .join("");
      return `
        <div class="mobile-year-block">
          <h3>${esc(y.name)}</h3>
          <div class="mobile-sem-row">${sems || '<p class="mobile-hint">No semesters yet.</p>'}</div>
        </div>`;
    })
    .join("");

  document.getElementById("coursesOutput").innerHTML = `
    <button type="button" class="mobile-back" data-mobile-back="program">← Programs</button>
    <div class="mobile-section-label">${esc(prog.name)}</div>
    ${years}`;
}

function renderMobileList() {
  hideExtra();
  const prog = findProgram(AppState.currentProg);
  const year = prog?.years.find((y) => idsEqual(y.id, AppState.currentYear));
  const sem = year?.sems.find((s) => idsEqual(s.id, AppState.currentSem));
  if (!prog || !year || !sem) {
    AppState.mobileStep = "year";
    renderMobileYearPicker();
    return;
  }

  const semKey = `${AppState.currentProg}:${AppState.currentYear}:${AppState.currentSem}`;
  const service = window.pickServices?.(1, `semester:${semKey}:${performance.now()}:${Math.random()}`)[0];
  const courseCards = (sem.courses || []).map((c) => _buildCourseCard(c));
  const cards = service ? window.intersperse?.(courseCards, [service], "semester") : courseCards;
  document.getElementById("coursesOutput").innerHTML = `
    ${chipsHtml([prog.name, year.name, sem.name], "year")}
    ${cards.length
      ? `<div class="courses-grid">${cards.join("")}</div>`
      : '<div class="empty">No courses in this semester — try another, or search.</div>'}`;
}

function renderMobileFavorites() {
  hideExtra();
  if (!window.isRegisteredStudent?.() && !AppState.adminLoggedIn) {
    document.getElementById("coursesOutput").innerHTML =
      chipsHtml(["My Courses"], "program") +
      '<div class="empty">Sign up to save courses here — it takes a name and a number.</div>';
    return;
  }
  if (AppState.favorites.size === 0) {
    document.getElementById("coursesOutput").innerHTML =
      chipsHtml(["My Courses"], "program") +
      '<div class="empty">No favorites yet — tap ★ on a course to save it here.</div>';
    return;
  }

  const q = searchQuery();
  const cards = [];
  AppState.dbPrograms.forEach((prog) => {
    prog.years.forEach((year) => {
      year.sems.forEach((sem) => {
        sem.courses.forEach((c) => {
          if (!AppState.favorites.has(String(c.id))) return;
          if (q && !c.name.toLowerCase().includes(q) && !c.code.toLowerCase().includes(q))
            return;
          cards.push(_buildCourseCard(c, { path: `${prog.name} · ${year.name} · ${sem.name}` }));
        });
      });
    });
  });

  document.getElementById("coursesOutput").innerHTML = `
    ${chipsHtml(["My Courses"], "program")}
    ${cards.length
      ? `<div class="courses-grid">${cards.join("")}</div>`
      : '<div class="empty">No matching favorites found.</div>'}`;
}

function renderMobileExtra() {
  AppState.currentProg = "extra";
  AppState.mobileStep = "list";
  showExtraOnly();
  window.renderExtra();
  const extra = document.getElementById("extraSection");
  if (!extra) return;
  extra.insertAdjacentHTML("afterbegin", chipsHtml(["Extra resources"], "program"));
  if (!extra.querySelector(".extra-section")) {
    extra.insertAdjacentHTML(
      "beforeend",
      '<div class="empty">No extra resources yet.</div>',
    );
  }
}

function renderMobileHome() {
  if (!isMobileView()) return false;

  const q = searchQuery();
  if (q) {
    renderMobileSearch(q);
    return true;
  }

  if (AppState.currentProg === "extra") {
    renderMobileExtra();
    return true;
  }
  if (AppState.currentProg === "community") {
    window.renderMobileCommunity?.();
    return true;
  }
  if (AppState.currentProg === "favorites") {
    renderMobileFavorites();
    return true;
  }
  if (AppState.mobileStep === "year" && isRealProgram(AppState.currentProg)) {
    renderMobileYearPicker();
    return true;
  }
  if (
    AppState.mobileStep === "list" &&
    isRealProgram(AppState.currentProg) &&
    AppState.currentYear !== "all" &&
    AppState.currentSem !== "all"
  ) {
    renderMobileList();
    return true;
  }

  AppState.mobileStep = "program";
  renderMobileProgramPicker();
  return true;
}

function selectMobileProg(id) {
  if (id === "favorites" && !window.requireStudent(() => selectMobileProg("favorites"))) {
    return;
  }

  AppState.currentProg = id;
  AppState.currentYear = "all";
  AppState.currentSem = "all";

  if (id === "all") {
    AppState.mobileStep = "program";
    renderMobileHome();
    return;
  }
  if (id === "extra") {
    renderMobileExtra();
    return;
  }
  if (id === "community") {
    AppState.mobileStep = "list";
    window.renderMobileCommunity?.();
    return;
  }
  if (id === "favorites") {
    AppState.mobileStep = "list";
    renderMobileFavorites();
    return;
  }

  AppState.mobileStep = "year";
  window.trackBrowse?.("year");
  renderMobileYearPicker();
}

function selectMobileSem(yearId, semId) {
  AppState.currentYear = coerceId(yearId);
  AppState.currentSem = coerceId(semId);
  AppState.mobileStep = "list";
  window.trackBrowse?.("list");
  renderMobileList();
}

function mobileBrowseBack(step) {
  if (step === "program") {
    AppState.currentProg = "all";
    AppState.currentYear = "all";
    AppState.currentSem = "all";
    AppState.mobileStep = "program";
    hideExtra();
    renderMobileProgramPicker();
    return;
  }
  if (step === "year") {
    AppState.mobileStep = "year";
    AppState.currentYear = "all";
    AppState.currentSem = "all";
    renderMobileYearPicker();
  }
}

function toggleCourseCard(courseId) {
  const card = document.getElementById(`course-card-${courseId}`);
  if (!card) return;
  const open = card.classList.contains("open");
  document.querySelectorAll(".course-card.open").forEach((el) => el.classList.remove("open"));
  if (!open) card.classList.add("open");
}

function initMobileHomeState() {
  if (!isMobileView()) return;
  AppState.currentProg = "all";
  AppState.currentYear = "all";
  AppState.currentSem = "all";
  AppState.mobileStep = "program";
}

function onMobileViewportChange() {
  if (document.getElementById("view-admin")?.classList.contains("active")) {
    window.renderAdminContent?.();
  }
  if (!document.getElementById("coursesOutput") || !AppState.dbPrograms.length) return;
  if (isMobileView()) {
    initMobileHomeState();
    const filterRow = document.querySelector(".filter-row");
    if (filterRow) filterRow.style.display = "none";
    renderMobileHome();
    return;
  }
  if (!AppState.currentProg) AppState.currentProg = "all";
  if (AppState.currentYear == null) AppState.currentYear = "all";
  if (AppState.currentSem == null) AppState.currentSem = "all";
  if (AppState.mobileStep === "program" || AppState.mobileStep === "year") {
    if (!isRealProgram(AppState.currentProg) && AppState.currentProg !== "community") AppState.currentProg = "all";
    AppState.currentYear = "all";
    AppState.currentSem = "all";
  }
  if (AppState.currentProg === "community") {
    window.selectCommunity?.();
    return;
  }
  window.selectProg(AppState.currentProg);
}

function onMobileHomeClick(e) {
  if (!isMobileView()) return;

  const progEl = e.target.closest("[data-mobile-prog]");
  if (progEl) {
    e.preventDefault();
    selectMobileProg(coerceId(progEl.dataset.mobileProg));
    return;
  }

  const semEl = e.target.closest("[data-mobile-sem]");
  if (semEl) {
    e.preventDefault();
    selectMobileSem(semEl.dataset.mobileYear, semEl.dataset.mobileSem);
    return;
  }

  const backEl = e.target.closest("[data-mobile-back]");
  if (backEl) {
    e.preventDefault();
    mobileBrowseBack(backEl.dataset.mobileBack);
    return;
  }

  if (e.target.closest(".fav-btn, .link-item, .copy-btn")) return;

  const extraTitle = e.target.closest(".extra-title");
  if (extraTitle) {
    const section = extraTitle.closest(".extra-section");
    if (section) {
      e.preventDefault();
      const open = section.classList.contains("open");
      document.querySelectorAll(".extra-section.open").forEach((el) => el.classList.remove("open"));
      if (!open) section.classList.add("open");
    }
    return;
  }

  const header = e.target.closest("[data-toggle-course]");
  if (header) {
    e.preventDefault();
    toggleCourseCard(header.dataset.toggleCourse);
    return;
  }

  const serviceHeader = e.target.closest("[data-toggle-service]");
  if (serviceHeader) {
    e.preventDefault();
    const card = serviceHeader.closest(".service-card");
    if (!card) return;
    const open = card.classList.contains("open");
    document.querySelectorAll(".course-card.open").forEach((el) => el.classList.remove("open"));
    if (!open) card.classList.add("open");
  }
}

function bindMobileHome() {
  const home = document.getElementById("view-home");
  if (home && !home.dataset.mobileBound) {
    home.addEventListener("click", onMobileHomeClick);
    home.dataset.mobileBound = "1";
  }
  window.matchMedia(MOBILE_MQ).addEventListener("change", onMobileViewportChange);
}

bindMobileHome();

window.renderMobileHome = renderMobileHome;
window.selectMobileProg = selectMobileProg;
window.initMobileHomeState = initMobileHomeState;
window.toggleCourseCard = toggleCourseCard;

export {
  renderMobileHome,
  selectMobileProg,
  initMobileHomeState,
  toggleCourseCard,
  onMobileViewportChange,
};
