// ===================== HOME RENDER =====================
function renderProgTabs() {
  document.getElementById("progTabs").innerHTML =
    `<button class="prog-tab ${AppState.currentProg === "all" ? "active" : ""}" onclick="selectProg('all')">All</button>` +
    AppState.dbPrograms
      .map(
        (p) =>
          `<button class="prog-tab ${p.id === AppState.currentProg ? "active" : ""}" onclick="selectProg(${p.id})">${esc(p.name)}</button>`,
      )
      .join("") +
    `<button class="prog-tab ${AppState.currentProg === "extra" ? "active" : ""}" onclick="selectProg('extra')">📦 Extra</button>` +
    `<button class="prog-tab fav-tab ${AppState.currentProg === "favorites" ? "active" : ""}" onclick="selectProg('favorites')">⭐ My Courses</button>`;
}

function renderYearFilters() {
  const prog = AppState.dbPrograms.find((p) => p.id === AppState.currentProg);
  if (!prog) return;
  document.getElementById("yearFilters").innerHTML =
    `<button class="filter-btn ${AppState.currentYear === "all" ? "active" : ""}" onclick="setYear('all')">All</button>` +
    prog.years
      .map(
        (y) =>
          `<button class="filter-btn ${AppState.currentYear === y.id ? "active" : ""}" onclick="setYear(${y.id})">${esc(y.name)}</button>`,
      )
      .join("");
}

function renderSemFilters() {
  const prog = AppState.dbPrograms.find((p) => p.id === AppState.currentProg);
  if (!prog) return;
  let sems = [];
  prog.years.forEach((y) => {
    if (AppState.currentYear === "all" || y.id === AppState.currentYear)
      y.sems.forEach((s) => {
        if (!sems.find((x) => x.id === s.id)) sems.push(s);
      });
  });
  document.getElementById("semFilters").innerHTML =
    `<button class="filter-btn ${AppState.currentSem === "all" ? "active" : ""}" onclick="setSem('all')">All</button>` +
    sems
      .map(
        (s) =>
          `<button class="filter-btn ${AppState.currentSem === s.id ? "active" : ""}" onclick="setSem(${s.id})">${esc(s.name)}</button>`,
      )
      .join("");
}

// ── Favorites view ─────────────────────────────────────────────────────────
function renderFavorites() {
  const q = document.getElementById("searchInput")?.value.toLowerCase().trim() || "";
  const favIds = AppState.favorites;
  let html = "";

  if (favIds.size === 0) {
    document.getElementById("coursesOutput").innerHTML =
      '<div class="empty">No favorites yet — click ★ on any course card to save it here.</div>';
    return;
  }

  AppState.dbPrograms.forEach((prog) => {
    let progHtml = "";
    prog.years.forEach((year) => {
      year.sems.forEach((sem) => {
        const filtered = sem.courses.filter(
          (c) =>
            favIds.has(String(c.id)) &&
            (!q || c.name.toLowerCase().includes(q) || c.code.toLowerCase().includes(q)),
        );
        if (!filtered.length) return;
        const cardsHtml = filtered.map(_buildCourseCard).join("");
        progHtml += `
          <div class="sem-block">
            <div class="sem-title">${esc(sem.name)}</div>
            <div class="courses-grid">${cardsHtml}</div>
          </div>`;
      });
    });
    if (progHtml) {
      html += `
        <div style="margin-bottom:32px;">
          <h3 style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:16px;">${esc(prog.name)}</h3>
          ${progHtml}
        </div>`;
    }
  });

  document.getElementById("coursesOutput").innerHTML =
    html || '<div class="empty">No matching favorites found.</div>';
}

function renderCourses() {
  if (AppState.currentProg === "all") {
    renderAllCourses();
    return;
  }
  if (AppState.currentProg === "favorites") {
    renderFavorites();
    return;
  }

  const prog = AppState.dbPrograms.find((p) => p.id === AppState.currentProg);
  if (!prog) return;

  const q = document.getElementById("searchInput").value.toLowerCase().trim();
  let html = "";

  prog.years.forEach((year) => {
    if (AppState.currentYear !== "all" && year.id !== AppState.currentYear) return;

    let yearHtml = "";

    year.sems.forEach((sem) => {
      if (AppState.currentSem !== "all" && sem.id !== AppState.currentSem) return;

      const filtered = sem.courses.filter(
        (c) =>
          !q ||
          c.name.toLowerCase().includes(q) ||
          c.code.toLowerCase().includes(q),
      );
      if (!filtered.length) return;

      const cardsHtml = filtered.map(_buildCourseCard).join("");

      yearHtml += `
        <div class="sem-block">
          <div class="sem-title">${esc(sem.name)}</div>
          <div class="courses-grid">${cardsHtml}</div>
        </div>`;
    });

    if (yearHtml) {
      html += `
        <div style="margin-bottom:32px;">
          <h3 style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:16px;">${esc(year.name)}</h3>
          ${yearHtml}
        </div>`;
    }
  });

  document.getElementById("coursesOutput").innerHTML =
    html || '<div class="empty">No courses found.</div>';
}

// ── All programs view ──────────────────────────────────────────────────────
function renderAllCourses() {
  const q = document.getElementById("searchInput").value.toLowerCase().trim();
  let html = "";

  AppState.dbPrograms.forEach((prog) => {
    let progHtml = "";
    prog.years.forEach((year) => {
      let yearHtml = "";
      year.sems.forEach((sem) => {
        const filtered = sem.courses.filter(
          (c) =>
            !q ||
            c.name.toLowerCase().includes(q) ||
            c.code.toLowerCase().includes(q),
        );
        if (!filtered.length) return;

        const cardsHtml = filtered.map(_buildCourseCard).join("");
        yearHtml += `
          <div class="sem-block">
            <div class="sem-title">${esc(sem.name)}</div>
            <div class="courses-grid">${cardsHtml}</div>
          </div>`;
      });

      if (yearHtml) {
        progHtml += `
          <div style="margin-bottom:32px;">
            <h3 style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:16px;">${esc(year.name)}</h3>
            ${yearHtml}
          </div>`;
      }
    });

    if (progHtml) {
      html += `
        <div style="margin-bottom:40px;padding-bottom:20px;border-bottom:1px solid var(--border);">
          <h2 style="font-size:1.1rem;font-weight:800;color:var(--text);margin-bottom:20px;padding-bottom:10px;border-bottom:2px solid var(--accent);">${esc(prog.name)}</h2>
          ${progHtml}
        </div>`;
    }
  });

  document.getElementById("coursesOutput").innerHTML = html;
}

function renderExtra() {
  const q =
    document.getElementById("searchInput")?.value.toLowerCase().trim() || "";

  const filtered = q
    ? AppState.dbExtra.filter(
        (r) =>
          r.title.toLowerCase().includes(q) ||
          r.links.some((l) => l.label.toLowerCase().includes(q)),
      )
    : AppState.dbExtra;

  document.getElementById("extraSection").innerHTML = filtered.length
    ? `
    <h3 style="font-size:.85rem;font-weight:700;text-transform:uppercase;letter-spacing:2px;color:var(--muted);margin-bottom:16px;">📦 Extra Resources</h3>
    <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:14px;">
      ${filtered
        .map(
          (r) => `
        <div class="extra-section">
          <div class="extra-title"><span>${r.icon}</span>${esc(r.title)}</div>
          <div class="links-list">
            ${r.links
              .map(
                (l) => `
              <a class="link-item"
                 data-url="${esc(l.url)}"
                 onclick="confirmLink(null, '${esc(l.url)}'); return false;"
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
              .join("")}
          </div>
        </div>`,
        )
        .join("")}
    </div>`
    : "";
}

function selectProg(id) {
  AppState.currentProg = id;
  AppState.currentYear = "all";
  AppState.currentSem = "all";
  renderProgTabs();
  if (id === "extra") {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("coursesOutput").style.display = "none";
    document.getElementById("extraSection").style.display = "";
    renderExtra();
  } else if (id === "all") {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("coursesOutput").style.display = "";
    document.getElementById("extraSection").style.display = "";
    renderCourses();
    renderExtra();
  } else if (id === "favorites") {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("coursesOutput").style.display = "";
    document.getElementById("extraSection").style.display = "none";
    renderFavorites();
  } else {
    document.querySelector(".filter-row").style.display = "";
    document.getElementById("coursesOutput").style.display = "";
    document.getElementById("extraSection").style.display = "none";
    renderYearFilters();
    renderSemFilters();
    renderCourses();
  }
}

function setYear(y) {
  AppState.currentYear = y;
  AppState.currentSem = "all";
  renderYearFilters();
  renderSemFilters();
  renderCourses();
}
function setSem(s) {
  AppState.currentSem = s;
  renderSemFilters();
  renderCourses();
}