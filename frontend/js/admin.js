
// ===================== ADMIN AUTH =====================
async function checkLogin() {
  const email = document.getElementById("adminEmail").value.trim();
  const pass = document.getElementById("adminPass").value;
  document.getElementById("loginErr").textContent = "";
  const btn = document.querySelector(".login-wrap .btn-primary");
  setBtnLoading(btn, true, "Logging in…");
  try {
    AppState.sbToken = await sbAuth(email, pass);
    AppState.adminLoggedIn = true;
    document.getElementById("adminPass").value = "";
    showView("admin");
  } catch (e) {
    document.getElementById("loginErr").textContent = e.message;
  } finally {
    setBtnLoading(btn, false);
  }
}

async function logout() {
  await sbLogout();
  AppState.adminLoggedIn = false;
  showView("home");
}

// ===================== ADMIN TABS =====================
const ADMIN_PAGE_SIZE = 25;
const AdminPager = {
  reports: { page: 0, hasNext: false },
  contributions: { page: 0, hasNext: false },
  feedback: { page: 0, hasNext: false },
};

function _setAdminPage(tab, page) {
  if (!AdminPager[tab]) return;
  AdminPager[tab].page = Math.max(0, page);
}

function _renderAdminPager(tab, rerenderFnName) {
  const pager = AdminPager[tab];
  if (!pager) return "";
  const pageNum = pager.page + 1;
  return `<div style="display:flex;gap:8px;align-items:center;justify-content:flex-end;margin-top:12px;">
    <button class="action-btn" ${pager.page === 0 ? "disabled" : ""} onclick="adminSetPage('${tab}', -1, '${rerenderFnName}')">← Prev</button>
    <span style="font-size:.85rem;color:var(--muted);">Page ${pageNum}</span>
    <button class="action-btn" ${pager.hasNext ? "" : "disabled"} onclick="adminSetPage('${tab}', 1, '${rerenderFnName}')">Next →</button>
  </div>`;
}

function adminSetPage(tab, delta, rerenderFnName) {
  if (!AdminPager[tab]) return;
  _setAdminPage(tab, AdminPager[tab].page + delta);
  if (typeof window[rerenderFnName] === "function") window[rerenderFnName]();
}

function adminTab(t) {
  AppState.currentAdminTab = t;
  AppState.adminSearch = "";
  AppState.adminFilterProg = "all";
  AppState.adminFilterYear = "all";
  AppState.adminFilterSem = "all";
  if (AdminPager[t]) _setAdminPage(t, 0);
  if (t === "feedback" && typeof window.resetAdminFeedbackPage === "function") {
    window.resetAdminFeedbackPage();
  }
  document.querySelectorAll(".admin-tab").forEach((b) => {
    b.classList.toggle("active", b.dataset.adminTab === t);
  });
  renderAdminContent();
}

function renderAdminContent() {
  loadReportsBadges();
  if (AppState.currentAdminTab === "courses") renderAdminCourses();
  else if (AppState.currentAdminTab === "extra") renderAdminExtra();
  else if (AppState.currentAdminTab === "feedback") renderAdminFeedback();
  else if (AppState.currentAdminTab === "reports") renderAdminReports();
  else if (AppState.currentAdminTab === "contributions") renderAdminContributions();
  else renderAdminAnalytics();
}

function _refocusSearch() {
  const s = document.querySelector("#adminContent .admin-search");
  if (s) {
    s.focus();
    s.setSelectionRange(s.value.length, s.value.length);
  }
}

// ===================== ANALYTICS =====================
async function renderAdminAnalytics() {
  document.getElementById("adminContent").innerHTML = getAdminAnalyticsSkeleton();
  try {
    const [views, clicks] = await Promise.all([
      sb("page_views", "GET", null, null, "id,visited_at"),
      sb("link_clicks", "GET", null, null, "id,link_id,clicked_at").catch(() => []),
    ]);

    const now = new Date();
    const rangeMs = parseInt(AppState.analyticsRange) * 24 * 60 * 60 * 1000;
    const cutoff = new Date(now - rangeMs);

    const inRange = views.filter((v) => new Date(v.visited_at) >= cutoff);
    const totalAll = views.length;
    const totalRange = inRange.length;
    const todayStr = now.toISOString().slice(0, 10);
    const todayCount = views.filter((v) => v.visited_at.slice(0, 10) === todayStr).length;

    const weekCutoff = new Date(now - 7 * 24 * 60 * 60 * 1000);
    const weekCount = views.filter((v) => new Date(v.visited_at) >= weekCutoff).length;

    const dayMap = {};
    inRange.forEach((v) => {
      const d = v.visited_at.slice(0, 10);
      dayMap[d] = (dayMap[d] || 0) + 1;
    });

    const days = [];
    for (let i = parseInt(AppState.analyticsRange) - 1; i >= 0; i--) {
      const d = new Date(now - i * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
      days.push({ date: d, count: dayMap[d] || 0 });
    }

    const maxCount = Math.max(...days.map((d) => d.count), 1);
    function fmtDay(dateStr) {
      const d = new Date(dateStr + "T00:00:00");
      return d.toLocaleDateString("en", { month: "short", day: "numeric" });
    }

    const labelStep = AppState.analyticsRange === "7" ? 1 : AppState.analyticsRange === "30" ? 5 : 15;
    const barsHtml = days
      .map((d, i) => {
        const pct = Math.round((d.count / maxCount) * 100);
        const showLabel = i % labelStep === 0 || i === days.length - 1;
        return `<div class="bar-col"><div class="bar-val" style="visibility:${d.count > 0 ? "visible" : "hidden"}">${d.count || ""}</div><div class="bar-fill" style="height:${Math.max(pct, d.count > 0 ? 4 : 0)}%;background:${d.date === todayStr ? "var(--accent2)" : "var(--accent)"}"></div><div class="bar-label">${showLabel ? fmtDay(d.date) : ""}</div></div>`;
      })
      .join("");

    const rangeButtons = ["7", "30", "90"]
      .map((r) => `<button class="filter-btn ${AppState.analyticsRange === r ? "active" : ""}" onclick="AppState.analyticsRange='${r}';renderAdminAnalytics()">${r} days</button>`)
      .join("");

    // Calculate top clicked links
    const clicksInRange = clicks.filter((c) => new Date(c.clicked_at) >= cutoff);
    const clickMap = {};
    clicksInRange.forEach((c) => {
      if (c.link_id) clickMap[c.link_id] = (clickMap[c.link_id] || 0) + 1;
    });

    const topLinks = Object.entries(clickMap)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([linkId, count]) => {
        // Resolve link label and course
        let info = { label: "Unknown Link", courseName: "Unknown Course" };
        AppState.dbPrograms.forEach(p => p.years.forEach(y => y.sems.forEach(s => s.courses.forEach(c => c.links.forEach(l => {
          if (l.id == linkId) info = { label: l.label, courseName: c.name };
        })))));
        AppState.dbExtra.forEach(r => r.links.forEach(l => {
          if (l.id == linkId) info = { label: l.label, courseName: r.title };
        }));
        return `<li><strong>${count}</strong> clicks: ${esc(info.label)} <span style="color:var(--muted);font-size:0.8rem">(${esc(info.courseName)})</span></li>`;
      })
      .join("");

    const topLinksHtml = topLinks ? `<ul style="list-style:none;padding:0;margin-top:16px;">${topLinks}</ul>` : `<div style="color:var(--muted);margin-top:16px;font-size:0.9rem;">No click data in range.</div>`;

    document.getElementById("adminContent").innerHTML = `
      <div class="stat-grid">
          <div class="stat-card"><div class="stat-val">${totalAll.toLocaleString()}</div><div class="stat-label">Total visits all time</div></div>
          <div class="stat-card"><div class="stat-val">${todayCount.toLocaleString()}</div><div class="stat-label">Visits today</div></div>
          <div class="stat-card"><div class="stat-val">${weekCount.toLocaleString()}</div><div class="stat-label">Visits this week</div></div>
          <div class="stat-card"><div class="stat-val">${totalRange.toLocaleString()}</div><div class="stat-label">Visits in range</div></div>
      </div>
      <div class="chart-wrap">
          <div class="chart-title">Daily visits — <span style="color:var(--accent2);">■</span> today</div>
          <div class="analytics-range">${rangeButtons}</div>
          <div class="bar-chart">${barsHtml}</div>
      </div>
      <div class="chart-wrap" style="margin-top:20px;">
          <div class="chart-title">🔥 Top Clicked Links (in range)</div>
          ${topLinksHtml}
      </div>
      <p style="font-size:.78rem;color:var(--muted);margin-top:8px;">Each visit counted once per browser session.</p>`;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ Could not load analytics: ${e.message}</div>`;
  }
}

// ===================== ADMIN COURSES =====================
function renderAdminCourses() {
  const q = AppState.adminSearch.toLowerCase();

  const progBtns =
    `<button class="filter-btn ${AppState.adminFilterProg === "all" ? "active" : ""}" onclick="AppState.adminFilterProg='all';AppState.adminFilterYear='all';AppState.adminFilterSem='all';renderAdminCourses()">All</button>` +
    AppState.dbPrograms
      .map((p) => `<button class="filter-btn ${AppState.adminFilterProg === p.id ? "active" : ""}" onclick="AppState.adminFilterProg=${p.id};AppState.adminFilterYear='all';AppState.adminFilterSem='all';renderAdminCourses()">${esc(p.name)}</button>`)
      .join("");

  const activeProg = AppState.dbPrograms.find((p) => p.id === AppState.adminFilterProg);

  let yearBtns = "";
  if (activeProg) {
    yearBtns =
      `<button class="filter-btn ${AppState.adminFilterYear === "all" ? "active" : ""}" onclick="AppState.adminFilterYear='all';AppState.adminFilterSem='all';renderAdminCourses()">All</button>` +
      activeProg.years.map((y) => `<button class="filter-btn ${AppState.adminFilterYear === y.id ? "active" : ""}" onclick="AppState.adminFilterYear=${y.id};AppState.adminFilterSem='all';renderAdminCourses()">${esc(y.name)}</button>`).join("");
  }

  let semBtns = "";
  if (activeProg) {
    let sems = [];
    activeProg.years.forEach((y) => {
      if (AppState.adminFilterYear === "all" || y.id === AppState.adminFilterYear)
        y.sems.forEach((s) => {
          if (!sems.find((x) => x.id === s.id)) sems.push(s);
        });
    });
    semBtns =
      `<button class="filter-btn ${AppState.adminFilterSem === "all" ? "active" : ""}" onclick="AppState.adminFilterSem='all';renderAdminCourses()">All</button>` +
      sems.map((s) => `<button class="filter-btn ${AppState.adminFilterSem === s.id ? "active" : ""}" onclick="AppState.adminFilterSem=${s.id};renderAdminCourses()">${esc(s.name)}</button>`).join("");
  }

  let html = `
    <input class="admin-search" placeholder="🔍 Search courses…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;renderAdminCourses()"/>
    <div style="margin-bottom:6px;">
        <div style="font-size:.7rem;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin-bottom:4px;">Program</div>
        <div class="filters" style="flex-wrap:wrap;">${progBtns}</div>
    </div>
    ${activeProg
      ? `
    <div style="margin-bottom:6px;">
        <div style="font-size:.7rem;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin-bottom:4px;">Year</div>
        <div class="filters" style="flex-wrap:wrap;">${yearBtns}</div>
    </div>
    <div style="margin-bottom:16px;">
        <div style="font-size:.7rem;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin-bottom:4px;">Semester</div>
        <div class="filters" style="flex-wrap:wrap;">${semBtns}</div>
    </div>`
      : `<div style="margin-bottom:16px;"></div>`
    }`;

  AppState.dbPrograms.forEach((prog) => {
    if (AppState.adminFilterProg !== "all" && prog.id !== AppState.adminFilterProg) return;
    let progHtml = "";
    prog.years.forEach((year) => {
      if (AppState.adminFilterYear !== "all" && year.id !== AppState.adminFilterYear) return;
      year.sems.forEach((sem) => {
        if (AppState.adminFilterSem !== "all" && sem.id !== AppState.adminFilterSem) return;
        const filtered = sem.courses.filter(
          (c) => !q || c.name.toLowerCase().includes(q) || c.code.toLowerCase().includes(q),
        );
        if (!filtered.length) return;
        progHtml += `<div style="font-size:.78rem;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin:12px 0 8px;">${esc(year.name)} — ${esc(sem.name)}</div>`;
        filtered.forEach((c) => {
          progHtml += `<div style="background:var(--bg3);border:1px solid var(--border);border-radius:10px;padding:12px 16px;margin-bottom:8px;">
            <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;">
              <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;">
                <strong>${esc(c.name)}</strong>
                <span style="color:var(--accent);font-size:.78rem;">${esc(c.code)}</span>
                ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
              </div>
              <div class="action-btns">
                <button class="action-btn" onclick="toggleOptional(${c.id},${c.is_optional})">${c.is_optional ? "✅ Optional" : "⬜ Optional"}</button>
                <button class="action-btn" onclick="openEditCourseModal(${c.id})">✏️ Edit</button>
                <button class="action-btn" onclick="openAddLinkModal(${c.id})">+ Link</button>
                <button class="action-btn del" onclick="confirmAction('Delete this course and all its links?',()=>deleteCourse(${c.id}))">🗑</button>
              </div>
            </div>
            <div style="margin-top:10px;display:flex;flex-wrap:wrap;gap:6px;">
              ${c.links.length
              ? c.links
                .map(
                  (l) => `
                <div class="link-chip">
                  ${getLinkBadge(l.type)}<span>${esc(l.label)}</span>
                  ${getContentTypeChip(l.content_type)}
                  ${l.note ? `<span style="color:var(--muted)">(${esc(l.note)})</span>` : ""}
                  <button class="action-btn" style="padding:2px 7px;font-size:.7rem;" onclick="openEditLinkModal(${l.id},${c.id})">✏️</button>
                  <button class="action-btn del" style="padding:2px 7px;font-size:.7rem;" onclick="confirmDeleteLink(${l.id},${c.id})">✕</button>
                </div>`,
                )
                .join("")
              : '<span style="font-size:.78rem;color:var(--muted);">No links</span>'
            }
            </div>
          </div>`;
        });
      });
    });
    if (progHtml)
      html += `<div style="margin-bottom:28px;"><div style="font-size:1rem;font-weight:700;color:var(--accent);margin-bottom:12px;">${esc(prog.name)}</div>${progHtml}</div>`;
  });

  document.getElementById("adminContent").innerHTML = html;
  _refocusSearch();
}

// ===================== ADMIN EXTRA =====================
function renderAdminExtra() {
  const q = AppState.adminSearch.toLowerCase();
  let html = `<input class="admin-search" placeholder="🔍 Search extra resources…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;renderAdminExtra()"/>
  <button class="btn btn-primary" style="margin-bottom:16px;" onclick="openAddExtraSectionModal()">+ Add Section</button>`;
  AppState.dbExtra.forEach((r) => {
    if (q && !r.title.toLowerCase().includes(q)) return;
    html += `<div style="background:var(--bg3);border:1px solid var(--border);border-radius:10px;padding:14px 16px;margin-bottom:10px;">
      <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:10px;">
        <div style="font-weight:700;">${r.icon} ${esc(r.title)}</div>
        <div class="action-btns">
          <button class="action-btn" onclick="openEditExtraSectionModal(${r.id})">✏️ Edit</button>
          <button class="action-btn" onclick="openAddExtraLinkModal(${r.id})">+ Link</button>
          <button class="action-btn del" onclick="confirmAction('Delete this section and all its links?',()=>deleteExtraSection(${r.id}))">🗑 Delete</button>
        </div>
      </div>
      <div style="display:flex;flex-wrap:wrap;gap:6px;">
        ${r.links.length
        ? r.links
          .map(
            (l) => `
          <div class="link-chip">
            ${getLinkBadge(l.type)}<span>${esc(l.label)}</span>
            ${getContentTypeChip(l.content_type)}
            ${l.note ? `<span style="color:var(--muted)">(${esc(l.note)})</span>` : ""}
            <button class="action-btn" style="padding:2px 7px;font-size:.7rem;" onclick="openEditExtraLinkModal(${l.id},${r.id})">✏️</button>
            <button class="action-btn del" style="padding:2px 7px;font-size:.7rem;" onclick="confirmAction('Remove this link?',()=>deleteExtraLink(${l.id},${r.id}))">✕</button>
          </div>`,
          )
          .join("")
        : '<span style="font-size:.78rem;color:var(--muted);">No links</span>'
      }
      </div>
    </div>`;
  });
  document.getElementById("adminContent").innerHTML = html;
}

// ===================== ADMIN REPORTS =====================
async function renderAdminReports() {
  document.getElementById("adminContent").innerHTML = getAdminTableSkeleton();
  const q = AppState.adminSearch.trim();
  const page = AdminPager.reports.page;
  const offset = page * ADMIN_PAGE_SIZE;
  try {
    const reports = await sb(`reports?limit=${ADMIN_PAGE_SIZE}&offset=${offset}&q=${encodeURIComponent(q)}`, "GET");
    if (page > 0 && reports.length === 0) {
      _setAdminPage("reports", page - 1);
      renderAdminReports();
      return;
    }
    AdminPager.reports.hasNext = reports.length === ADMIN_PAGE_SIZE;
    let html = `<input class="admin-search" placeholder="🔍 Search reports…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;_setAdminPage('reports',0);renderAdminReports()"/>`;
    if (!reports.length) {
      document.getElementById("adminContent").innerHTML = html + '<div class="empty">No reports yet.</div>';
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Course</th><th>Link</th><th>Issue</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    reports.forEach((r) => {
      html += `<tr><td>${esc(r.course_name)}</td><td style="max-width:140px;word-break:break-all;font-size:.75rem;">${esc(r.link_url || "—")}</td><td>${esc(r.description)}</td>
        <td><span class="tag ${r.status === "open" ? "tag-open" : "tag-resolved"}">${r.status}</span></td>
        <td class="action-btns">
          <button class="action-btn" onclick="toggleReportStatus(${r.id},'${r.status}')">${r.status === "open" ? "✅ Resolve" : "↩ Reopen"}</button>
          <button class="action-btn del" onclick="confirmAction('Delete this report?',()=>deleteReport(${r.id}))">🗑</button>
        </td></tr>`;
    });
    html += "</tbody></table>";
    html += _renderAdminPager("reports", "renderAdminReports");
    document.getElementById("adminContent").innerHTML = html;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ ${e.message}</div>`;
  }
}

// ===================== ADMIN CONTRIBUTIONS =====================
async function renderAdminContributions() {
  document.getElementById("adminContent").innerHTML = getAdminTableSkeleton();
  const q = AppState.adminSearch.trim();
  const page = AdminPager.contributions.page;
  const offset = page * ADMIN_PAGE_SIZE;
  try {
    const contribs = await sb(`contributions?limit=${ADMIN_PAGE_SIZE}&offset=${offset}&q=${encodeURIComponent(q)}`, "GET");
    if (page > 0 && contribs.length === 0) {
      _setAdminPage("contributions", page - 1);
      renderAdminContributions();
      return;
    }
    AdminPager.contributions.hasNext = contribs.length === ADMIN_PAGE_SIZE;
    let html = `<input class="admin-search" placeholder="🔍 Search contributions…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;_setAdminPage('contributions',0);renderAdminContributions()"/>`;
    if (!contribs.length) {
      document.getElementById("adminContent").innerHTML = html + '<div class="empty">No contributions yet.</div>';
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Course</th><th>Link</th><th>Note</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    contribs.forEach((c) => {
      html += `<tr><td>${esc(c.course_name)}</td><td style="max-width:140px;word-break:break-all;font-size:.75rem;">${esc(c.link_url)}</td><td>${esc(c.note || "—")}</td>
        <td><span class="tag ${c.status === "pending" ? "tag-open" : "tag-resolved"}">${c.status}</span></td>
        <td class="action-btns">
          ${c.status === "pending" ? `<button class="action-btn" style="color:var(--success); border-color:var(--success)" onclick='openAutoApproveContribModal(${JSON.stringify(c).replace(/'/g, "&apos;")})'>✅ Approve & Add</button>` : `<button class="action-btn" disabled>Approved</button>`}
          <button class="action-btn del" onclick="confirmAction('Reject this contribution?',()=>deleteContrib(${c.id}))">🗑 Reject</button>
        </td></tr>`;
    });
    html += "</tbody></table>";
    html += _renderAdminPager("contributions", "renderAdminContributions");
    document.getElementById("adminContent").innerHTML = html;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ ${e.message}</div>`;
  }
}

// Auto-Approve Flow
function openAutoApproveContribModal(c) {
  let suggestedCourseId = "";
  AppState.dbPrograms.forEach(p => p.years.forEach(y => y.sems.forEach(s => s.courses.forEach(course => {
    if (course.name.toLowerCase() === c.course_name.toLowerCase() || course.code.toLowerCase() === c.course_name.toLowerCase()) {
      suggestedCourseId = course.id;
    }
  }))));

  let courseOpts = "";
  AppState.dbPrograms.forEach(p => {
    courseOpts += `<optgroup label="${esc(p.name)}">`;
    p.years.forEach(y => y.sems.forEach(s => s.courses.forEach(course => {
      courseOpts += `<option value="${course.id}" ${course.id === suggestedCourseId ? "selected" : ""}>${esc(course.name)} (${esc(course.code)}) — ${esc(y.name)} › ${esc(s.name)}</option>`;
    })));
    courseOpts += `</optgroup>`;
  });

  let preType = c.link_type || "drive";
  let cleanNote = c.note || "";
  const match = cleanNote.match(/^\[Type:\s*([^\]]+)\]\s*(.*)$/i);
  if (match) {
    preType = match[1];
    cleanNote = match[2];
  }

  openModal(`<h2>✅ Approve Contribution</h2>
  <p style="color:var(--muted);font-size:0.9rem;margin-bottom:16px;">Review and add this link directly to the database.</p>
  <label>Target Course</label><select id="acCourse">${courseOpts}</select>
  <label>Type</label><select id="acType">${_linkTypeOptions(preType)}</select>
  <label>URL</label><input type="text" id="acUrl" value="${esc(c.link_url)}"/>
  <label>Label</label><input type="text" id="acLabel" value="Link"/>
  <label>Content Type(s)</label>${_contentTypeCheckboxes("", "acct")}
  <label>Note</label><input type="text" id="acNote" value="${esc(cleanNote)}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="applyAutoApproveContrib(${c.id})">Approve & Add</button></div>`);
}

async function applyAutoApproveContrib(contribId) {
  const courseId = parseInt(document.getElementById("acCourse").value);
  const url = document.getElementById("acUrl").value.trim();
  const type = document.getElementById("acType").value;
  const label = document.getElementById("acLabel").value.trim() || "Link";
  const note = document.getElementById("acNote").value.trim();
  const contentType = _readContentTypeCheckboxes("acct");

  if (!courseId || !url) { showToast("Course and URL required.", true); return; }

  const { siblings } = _findSharedCourses(courseId);

  if (siblings.length) {
    AppState._pendingLinkOp = { contribId, courseId, url, type, label, note, contentType, siblingCourseIds: siblings.map(s => s.id) };
    const list = siblings
      .map(s => `<li style="margin-bottom:4px;"><strong>${esc(s.name)}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${esc(s.prog)} › ${esc(s.year)} › ${esc(s.sem)}</span></li>`)
      .join("");
    openModal(`<h2>🔁 Shared Course</h2>
    <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${siblings.length}</strong> other course(s) share the same code:</p>
    <ul style="font-size:.83rem;margin-bottom:16px;padding-left:18px;list-style:disc;">${list}</ul>
    <p style="font-size:.85rem;margin-bottom:16px;">Add this link to all of them too?</p>
    <div class="modal-actions">
        <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
        <button class="btn btn-ghost" onclick="_applyAutoApproveWithSiblings(false)">Just this one</button>
        <button class="btn btn-primary" onclick="_applyAutoApproveWithSiblings(true)">All ${siblings.length + 1} courses</button>
    </div>`);
  } else {
    const btn = document.querySelector("#modalBox .btn-primary");
    setBtnLoading(btn, true, "Saving…");
    try {
      await sb("links", "POST", {
        course_id: courseId, type, url, label, note, content_type: contentType, display_order: _getNextDisplayOrder(courseId)
      });
      await sb(`contributions?id=eq.${contribId}`, "PATCH", { status: "approved" });
      closeModal(); _clearCache(); loadAll(); renderAdminContributions();
      showToast("Contribution approved and link added!");
    } catch (e) { showToast(e.message, true); }
    finally { setBtnLoading(btn, false); }
  }
}

async function _applyAutoApproveWithSiblings(addToAll) {
  const { contribId, courseId, url, type, label, note, contentType, siblingCourseIds } = AppState._pendingLinkOp;
  const btn = document.querySelector("#modalBox .btn-primary");
  setBtnLoading(btn, true, "Saving…");
  try {
    await sb("links", "POST", { course_id: courseId, type, url, label, note, content_type: contentType, display_order: _getNextDisplayOrder(courseId) });
    if (addToAll && siblingCourseIds.length)
      await Promise.all(siblingCourseIds.map(sid => sb("links", "POST", { course_id: sid, type, url, label, note, content_type: contentType, display_order: _getNextDisplayOrder(sid) })));
    await sb(`contributions?id=eq.${contribId}`, "PATCH", { status: "approved" });
    closeModal(); _clearCache(); loadAll(); renderAdminContributions();
    showToast(addToAll ? `Link added to all ${siblingCourseIds.length + 1} courses!` : "Contribution approved and link added!");
  } catch (e) { showToast(e.message, true); }
  finally { AppState._pendingLinkOp = null; setBtnLoading(btn, false); }
}

// ===================== DB MUTATIONS =====================
async function deleteCourse(id) {
  try {
    await sb(`courses?id=eq.${id}`, "DELETE");
    _clearCache();
    loadAll();
    showToast("Course deleted.");
  } catch (e) { showToast(e.message, true); }
}
async function toggleOptional(id, current) {
  try {
    await sb(`courses?id=eq.${id}`, "PATCH", { is_optional: !current });
    _clearCache();
    loadAll();
    renderAdminCourses();
    showToast(current ? "Marked as required." : "Marked as optional.");
  } catch (e) { showToast(e.message, true); }
}
function confirmDeleteLink(linkId, courseId) {
  let linkUrl = null;
  AppState.dbPrograms.forEach((p) => p.years.forEach((y) => y.sems.forEach((s) => s.courses.forEach((c) => c.links.forEach((lk) => { if (lk.id === linkId) linkUrl = lk.url; })))));
  const { siblings } = _findSharedCourses(courseId);
  const matchingLinks = [];
  siblings.forEach((sib) => sib.links.forEach((lk) => { if (lk.url === linkUrl) matchingLinks.push({ id: lk.id, sibName: sib.name, prog: sib.prog, year: sib.year, sem: sib.sem }); }));

  AppState._pendingLinkOp = { linkId, matchingLinkIds: matchingLinks.map((m) => m.id) };

  if (matchingLinks.length) {
    const list = matchingLinks.map((m) => `<li style="margin-bottom:4px;"><strong>${esc(m.sibName)}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${esc(m.prog)} › ${esc(m.year)} › ${esc(m.sem)}</span></li>`).join("");
    openModal(`<h2>🗑 Delete Link</h2>
    <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${matchingLinks.length}</strong> sibling course(s) have the same link:</p>
    <ul style="font-size:.83rem;margin-bottom:16px;padding-left:18px;list-style:disc;">${list}</ul>
    <p style="font-size:.85rem;margin-bottom:16px;">Remove from all of them too?</p>
    <div class="modal-actions">
        <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
        <button class="btn btn-ghost" onclick="applyDeleteLink(false)">Just this one</button>
        <button class="btn" style="background:var(--danger);color:#fff;" onclick="applyDeleteLink(true)">All ${matchingLinks.length + 1} links</button>
    </div>`);
  } else {
    openModal(`<h2>⚠️ Confirm</h2>
    <p style="color:var(--muted);font-size:.9rem;margin-top:8px;">Remove this link?</p>
    <div class="modal-actions">
        <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
        <button class="btn" style="background:var(--danger);color:#fff;" onclick="applyDeleteLink(false)">Delete</button>
    </div>`);
  }
}
async function applyDeleteLink(deleteAll) {
  const { linkId, matchingLinkIds } = AppState._pendingLinkOp;
  try {
    await sb(`links?id=eq.${linkId}`, "DELETE");
    if (deleteAll && matchingLinkIds.length) await Promise.all(matchingLinkIds.map((mid) => sb(`links?id=eq.${mid}`, "DELETE")));
    closeModal();
    _clearCache();
    loadAll();
    renderAdminCourses();
    showToast(deleteAll ? `Removed ${matchingLinkIds.length + 1} links!` : "Link removed.");
  } catch (e) { showToast(e.message, true); } finally { AppState._pendingLinkOp = null; }
}
async function deleteExtraSection(id) {
  try {
    await sb(`extra_sections?id=eq.${id}`, "DELETE");
    _clearCache();
    loadAll();
    renderAdminExtra();
    showToast("Section deleted.");
  } catch (e) { showToast(e.message, true); }
}
async function deleteExtraLink(id) {
  try {
    await sb(`extra_links?id=eq.${id}`, "DELETE");
    _clearCache();
    loadAll();
    renderAdminExtra();
    showToast("Link removed.");
  } catch (e) { showToast(e.message, true); }
}
async function toggleReportStatus(id, status) {
  try {
    await sb(`reports?id=eq.${id}`, "PATCH", { status: status === "open" ? "resolved" : "open" });
    renderAdminReports();
    loadReportsBadges();
  } catch (e) { showToast(e.message, true); }
}
async function deleteReport(id) {
  try {
    await sb(`reports?id=eq.${id}`, "DELETE");
    renderAdminReports();
    loadReportsBadges();
    showToast("Report deleted.");
  } catch (e) { showToast(e.message, true); }
}
async function deleteContrib(id) {
  try {
    await sb(`contributions?id=eq.${id}`, "DELETE");
    renderAdminContributions();
    loadReportsBadges();
    showToast("Contribution rejected.");
  } catch (e) { showToast(e.message, true); }
}

// Setup badge auto-refresh
setInterval(() => {
  if (AppState.adminLoggedIn) loadReportsBadges();
}, 30000);

window.checkLogin = checkLogin;
window.logout = logout;
window.adminTab = adminTab;
window.renderAdminContent = renderAdminContent;
window.loadReportsBadges = loadReportsBadges;
window.adminSetPage = adminSetPage;
window.renderAdminReports = renderAdminReports;
window.renderAdminContributions = renderAdminContributions;
window.ADMIN_PAGE_SIZE = ADMIN_PAGE_SIZE;
