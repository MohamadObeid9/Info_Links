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
}

// Natural sort: extract trailing number from label for numeric ordering
function _naturalLinkSort(a, b) {
  const numA = parseInt((a.label || "").match(/(\d+)\s*$/)?.[1]) || 0;
  const numB = parseInt((b.label || "").match(/(\d+)\s*$/)?.[1]) || 0;
  if (numA !== numB) return numA - numB;
  return (a.display_order || 0) - (b.display_order || 0);
}

function _renderAfterLoad() {
  if (!AppState.currentProg) AppState.currentProg = "all";
  document.getElementById("extraSection").style.display = "none";
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
  const isFirstLoad = !document.getElementById("coursesOutput").dataset.loaded;
  if (isFirstLoad) {
    showSkeleton();
  } else {
    document.getElementById("coursesOutput").innerHTML =
      '<div class="loader"><div class="spinner"></div> Loading…</div>';
  }
  document.getElementById("extraSection").innerHTML = "";
  try {
    // Admins always get fresh data; visitors use the 1-hour cache.
    const cached = AppState.adminLoggedIn ? null : _loadCache();

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

    // Cache miss — fetch all tables in parallel, including content_type
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

    if (!AppState.adminLoggedIn) {
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
    document.getElementById("coursesOutput").dataset.loaded = "1";
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
  if (AppState.currentProg === "extra") renderExtra();
  else if (AppState.currentProg === "all") {
    renderCourses();
    renderExtra();
  } else renderCourses();
}
