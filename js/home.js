// ===================== HOME RENDER =====================
function renderProgTabs() {
  document.getElementById("progTabs").innerHTML =
    `<button class="prog-tab ${currentProg === "all" ? "active" : ""}" onclick="selectProg('all')">All</button>` +
    dbPrograms
      .map(
        (p) =>
          `<button class="prog-tab ${p.id === currentProg ? "active" : ""}" onclick="selectProg(${p.id})">${esc(p.name)}</button>`,
      )
      .join("") +
    `<button class="prog-tab ${currentProg === "extra" ? "active" : ""}" onclick="selectProg('extra')">📦 Extra Resources</button>`;
}

function renderYearFilters() {
  const prog = dbPrograms.find((p) => p.id === currentProg);
  if (!prog) return;
  document.getElementById("yearFilters").innerHTML =
    `<button class="filter-btn ${currentYear === "all" ? "active" : ""}" onclick="setYear('all')">All</button>` +
    prog.years
      .map(
        (y) =>
          `<button class="filter-btn ${currentYear === y.id ? "active" : ""}" onclick="setYear(${y.id})">${esc(y.name)}</button>`,
      )
      .join("");
}

function renderSemFilters() {
  const prog = dbPrograms.find((p) => p.id === currentProg);
  if (!prog) return;
  let sems = [];
  prog.years.forEach((y) => {
    if (currentYear === "all" || y.id === currentYear)
      y.sems.forEach((s) => {
        if (!sems.find((x) => x.id === s.id)) sems.push(s);
      });
  });
  document.getElementById("semFilters").innerHTML =
    `<button class="filter-btn ${currentSem === "all" ? "active" : ""}" onclick="setSem('all')">All</button>` +
    sems
      .map(
        (s) =>
          `<button class="filter-btn ${currentSem === s.id ? "active" : ""}" onclick="setSem(${s.id})">${esc(s.name)}</button>`,
      )
      .join("");
}

function renderCourses() {
  if (currentProg === "all") {
    renderAllCourses();
    return;
  }
  const prog = dbPrograms.find((p) => p.id === currentProg);
  if (!prog) return;

  const q = document.getElementById("searchInput").value.toLowerCase().trim();
  let html = "";

  prog.years.forEach((year) => {
    if (currentYear !== "all" && year.id !== currentYear) return;

    let yearHtml = "";

    year.sems.forEach((sem) => {
      if (currentSem !== "all" && sem.id !== currentSem) return;

      const filtered = sem.courses.filter(
        (c) =>
          !q ||
          c.name.toLowerCase().includes(q) ||
          c.code.toLowerCase().includes(q),
      );
      if (!filtered.length) return;

      const cardsHtml = filtered
        .map((c) => {
          const linksHtml = c.links.length
            ? c.links
              .map(
                (l) => `
                                    <a class="link-item" onclick="confirmLink('${l.url}'); return false;" href="#">
                                        ${getLinkBadge(l.type)}
                                        <span class="link-label">${esc(l.label)}</span>
                                        ${l.note ? `<span class="link-note">${esc(l.note)}</span>` : ""}
                                    </a>`,
              )
              .join("")
            : '<span class="no-links">No links yet — contribute!</span>';

          return `
                                <div class="course-card">
                                    <div class="course-header">
                                        <div class="course-name">${esc(c.name)}</div>
                                        <div style="display:flex;align-items:center;gap:6px;">
                                            ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
                                            <div class="course-code">${esc(c.code)}</div>
                                        </div>
                                    </div>
                                    <div class="links-list">${linksHtml}</div>
                                </div>`;
        })
        .join("");

      yearHtml += `
                            <div class="sem-block">
                                <div class="sem-title">${sem.name}</div>
                                <div class="courses-grid">${cardsHtml}</div>
                            </div>`;
    });

    if (yearHtml) {
      html += `
                            <div style="margin-bottom:32px;">
                                <h3 style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:16px;">${year.name}</h3>
                                ${yearHtml}
                            </div>`;
    }
  });

  document.getElementById("coursesOutput").innerHTML =
    html || '<div class="empty">No courses found.</div>';
}

function renderExtra() {
  const q =
    document.getElementById("searchInput")?.value.toLowerCase().trim() || "";

  const filtered = q
    ? dbExtra.filter(
      (r) =>
        r.title.toLowerCase().includes(q) ||
        r.links.some((l) => l.label.toLowerCase().includes(q)),
    )
    : dbExtra;

  document.getElementById("extraSection").innerHTML = filtered.length
    ? `
    <h3 style="font-size:.85rem;font-weight:700;text-transform:uppercase;letter-spacing:2px;color:var(--muted);margin-bottom:16px;">📦 Extra Resources</h3>
    <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:14px;">
      ${filtered
      .map(
        (r) => `
        <div class="extra-section">
          <div class="extra-title"><span>${r.icon}</span>${r.title}</div>
          <div class="links-list">
            ${r.links
            .map(
              (l) => `
              <a class="link-item" onclick="confirmLink('${l.url}'); return false;" href="#">
                ${getLinkBadge(l.type)}<span class="link-label">${l.label}</span>${l.note ? `<span class="link-note">${l.note}</span>` : ""}
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
  currentProg = id;
  currentYear = "all";
  currentSem = "all";
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
  currentYear = y;
  currentSem = "all";
  renderYearFilters();
  renderSemFilters();
  renderCourses();
}
function setSem(s) {
  currentSem = s;
  renderSemFilters();
  renderCourses();
}