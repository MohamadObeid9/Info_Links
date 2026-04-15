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
  dbPrograms = programs.map((p) => ({
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
                links: links.filter((l) => l.course_id === c.id),
              })),
          })),
      })),
  }));

  dbExtra = extraSections.map((sec) => ({
    ...sec,
    links: extraLinks.filter((l) => l.section_id === sec.id),
  }));
}

function _renderAfterLoad() {
  if (!currentProg) currentProg = "all";
  document.getElementById("extraSection").style.display = "none";
  if (currentProg === "all") {
    document.querySelector(".filter-row").style.display = "none";
    document.getElementById("extraSection").style.display = "";
  }
  renderProgTabs();
  renderYearFilters();
  renderSemFilters();
  renderCourses();
  if (currentProg === "all") renderExtra();
}

async function loadAll() {
  document.getElementById("coursesOutput").innerHTML =
    '<div class="loader"><div class="spinner"></div> Loading courses…</div>';
  document.getElementById("extraSection").innerHTML = "";
  try {
    // Admins always get fresh data so they see their own changes.
    // Regular visitors use the cache (1 hour TTL) for instant loads.
    const cached = adminLoggedIn ? null : _loadCache();

    if (cached) {
      _buildTree(
        cached.programs,
        cached.years,
        cached.semesters,
        cached.courses,
        cached.links,
        cached.extra_sections,
        cached.extra_links,
      );
      _renderAfterLoad();
      return;
    }

    // Cache miss — fetch from Supabase
    const [
      programs,
      years,
      semesters,
      courses,
      links,
      extraSections,
      extraLinks,
    ] = await Promise.all([
      sb("programs", "GET", null, null, "*&order=display_order.asc"),
      sb("years", "GET", null, null, "*&order=display_order.asc"),
      sb("semesters", "GET", null, null, "*&order=display_order.asc"),
      sb("courses", "GET", null, null, "*&order=display_order.asc"),
      sb("links", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_sections", "GET", null, null, "*&order=display_order.asc"),
      sb("extra_links", "GET", null, null, "*&order=display_order.asc"),
    ]);

    // Save to cache (only for non-admin visitors)
    if (!adminLoggedIn) {
      _saveCache({
        programs,
        years,
        semesters,
        courses,
        links,
        extra_sections: extraSections,
        extra_links: extraLinks,
      });
    }

    _buildTree(
      programs,
      years,
      semesters,
      courses,
      links,
      extraSections,
      extraLinks,
    );
    _renderAfterLoad();
  } catch (e) {
    document.getElementById("coursesOutput").innerHTML =
      `<div class="empty">⚠️ Failed to load data: ${e.message}</div>`;
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
    cb.style.display = pendC ? "inline" : "none";
    cb.textContent = pendC;
    fb.style.display = newF ? "inline" : "none";
    fb.textContent = newF;
  } catch (e) {}
}

function onSearch() {
  if (currentProg === "extra") renderExtra();
  else if (currentProg === "all") {
    renderCourses();
    renderExtra();
  } else renderCourses();
}

function renderAllCourses() {
  const q = document.getElementById("searchInput").value.toLowerCase().trim();
  let html = "";

  dbPrograms.forEach((prog) => {
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

        const cardsHtml = filtered
          .map((c) => {
            const linksHtml = c.links.length
              ? c.links
                  .map(
                    (l) => `
                                    <a class="link-item" href="${l.url}" target="_blank" rel="noopener">
                                        ${getLinkBadge(l.type)}
                                        <span class="link-label">${l.label}</span>
                                        ${l.note ? `<span class="link-note">${l.note}</span>` : ""}
                                    </a>`,
                  )
                  .join("")
              : '<span class="no-links">No links yet — contribute!</span>';

            return `
                                <div class="course-card">
                                    <div class="course-header">
                                        <div class="course-name">${c.name}</div>
                                        <div style="display:flex;align-items:center;gap:6px;">
                                            ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
                                            <div class="course-code">${c.code}</div>
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
        progHtml += `
                            <div style="margin-bottom:32px;">
                                <h3 style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:16px;">${year.name}</h3>
                                ${yearHtml}
                            </div>`;
      }
    });

    if (progHtml) {
      html += `
                        <div style="margin-bottom:40px;padding-bottom:20px;border-bottom:1px solid var(--border);">
                            <h2 style="font-size:1.1rem;font-weight:800;color:var(--text);margin-bottom:20px;padding-bottom:10px;border-bottom:2px solid var(--accent);">${prog.name}</h2>
                            ${progHtml}
                        </div>`;
    }
  });

  document.getElementById("coursesOutput").innerHTML = html;
}