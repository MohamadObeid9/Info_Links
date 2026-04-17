// ===================== ADMIN AUTH =====================
async function checkLogin() {
  const email = document.getElementById("adminEmail").value.trim();
  const pass = document.getElementById("adminPass").value;
  document.getElementById("loginErr").textContent = "";
  try {
    sbToken = await sbAuth(email, pass);
    adminLoggedIn = true;
    document.getElementById("adminPass").value = "";
    showView("admin");
  } catch (e) {
    document.getElementById("loginErr").textContent = e.message;
  }
}

async function logout() {
  await sbLogout();
  adminLoggedIn = false;
  showView("home");
}

// ===================== ADMIN TABS =====================
function adminTab(t) {
  currentAdminTab = t;
  adminSearch = "";
  adminFilterProg = "all";
  adminFilterYear = "all";
  adminFilterSem = "all";
  document
    .querySelectorAll(".admin-tab")
    .forEach((b, i) =>
      b.classList.toggle(
        "active",
        ["courses", "extra", "feedback", "reports", "contributions", "analytics"][i] === t,
      ),
    );
  renderAdminContent();
}

function renderAdminContent() {
  loadReportsBadges();
  if (currentAdminTab === "courses") renderAdminCourses();
  else if (currentAdminTab === "extra") renderAdminExtra();
  else if (currentAdminTab === "feedback") renderAdminFeedback();
  else if (currentAdminTab === "reports") renderAdminReports();
  else if (currentAdminTab === "contributions") renderAdminContributions();
  else renderAdminAnalytics();
}

function _refocusSearch() {
  const s = document.querySelector("#adminContent .admin-search");
  if (s) {
    s.focus();
    s.setSelectionRange(s.value.length, s.value.length);
  }
}

// ===================== ADMIN COURSES =====================

// ===================== ANALYTICS =====================
async function renderAdminAnalytics() {
  document.getElementById("adminContent").innerHTML = getAdminAnalyticsSkeleton();
  try {
    const views = await sb("page_views", "GET", null, null, "id,visited_at");
    const now = new Date();
    const rangeMs = parseInt(analyticsRange) * 24 * 60 * 60 * 1000;
    const cutoff = new Date(now - rangeMs);
    const inRange = views.filter((v) => new Date(v.visited_at) >= cutoff);
    const totalAll = views.length;
    const totalRange = inRange.length;
    const todayStr = now.toISOString().slice(0, 10);
    const todayCount = views.filter(
      (v) => v.visited_at.slice(0, 10) === todayStr,
    ).length;
    const weekCutoff = new Date(now - 7 * 24 * 60 * 60 * 1000);
    const weekCount = views.filter(
      (v) => new Date(v.visited_at) >= weekCutoff,
    ).length;
    const dayMap = {};
    inRange.forEach((v) => {
      const d = v.visited_at.slice(0, 10);
      dayMap[d] = (dayMap[d] || 0) + 1;
    });
    const days = [];
    for (let i = parseInt(analyticsRange) - 1; i >= 0; i--) {
      const d = new Date(now - i * 24 * 60 * 60 * 1000)
        .toISOString()
        .slice(0, 10);
      days.push({ date: d, count: dayMap[d] || 0 });
    }
    const maxCount = Math.max(...days.map((d) => d.count), 1);
    function fmtDay(dateStr) {
      const d = new Date(dateStr + "T00:00:00");
      return d.toLocaleDateString("en", {
        month: "short",
        day: "numeric",
      });
    }
    const labelStep =
      analyticsRange === "7" ? 1 : analyticsRange === "30" ? 5 : 15;
    const barsHtml = days
      .map((d, i) => {
        const pct = Math.round((d.count / maxCount) * 100);
        const showLabel = i % labelStep === 0 || i === days.length - 1;
        return `<div class="bar-col"><div class="bar-val" style="visibility:${d.count > 0 ? "visible" : "hidden"}">${d.count || ""}</div><div class="bar-fill" style="height:${Math.max(pct, d.count > 0 ? 4 : 0)}%;background:${d.date === todayStr ? "var(--accent2)" : "var(--accent)"}"></div><div class="bar-label">${showLabel ? fmtDay(d.date) : ""}</div></div>`;
      })
      .join("");
    const rangeButtons = ["7", "30", "90"]
      .map(
        (r) =>
          `<button class="filter-btn ${analyticsRange === r ? "active" : ""}" onclick="analyticsRange='${r}';renderAdminAnalytics()">${r} days</button>`,
      )
      .join("");
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
                        <p style="font-size:.78rem;color:var(--muted);margin-top:8px;">Each visit counted once per browser session.</p>`;
  } catch (e) {
    document.getElementById("adminContent").innerHTML =
      `<div class="empty">⚠️ Could not load analytics: ${e.message}</div>`;
  }
}

function renderAdminCourses() {
  const q = adminSearch.toLowerCase();

  // Build filter bar HTML
  const progBtns =
    `<button class="filter-btn ${adminFilterProg === "all" ? "active" : ""}" onclick="adminFilterProg='all';adminFilterYear='all';adminFilterSem='all';renderAdminCourses()">All</button>` +
    dbPrograms
      .map(
        (p) =>
          `<button class="filter-btn ${adminFilterProg === p.id ? "active" : ""}" onclick="adminFilterProg=${p.id};adminFilterYear='all';adminFilterSem='all';renderAdminCourses()">${esc(p.name)}</button>`,
      )
      .join("");

  const activeProg = dbPrograms.find((p) => p.id === adminFilterProg);

  let yearBtns = "";
  if (activeProg) {
    yearBtns =
      `<button class="filter-btn ${adminFilterYear === "all" ? "active" : ""}" onclick="adminFilterYear='all';adminFilterSem='all';renderAdminCourses()">All</button>` +
      activeProg.years
        .map(
          (y) =>
            `<button class="filter-btn ${adminFilterYear === y.id ? "active" : ""}" onclick="adminFilterYear=${y.id};adminFilterSem='all';renderAdminCourses()">${esc(y.name)}</button>`,
        )
        .join("");
  }

  let semBtns = "";
  if (activeProg) {
    let sems = [];
    activeProg.years.forEach((y) => {
      if (adminFilterYear === "all" || y.id === adminFilterYear)
        y.sems.forEach((s) => {
          if (!sems.find((x) => x.id === s.id)) sems.push(s);
        });
    });
    semBtns =
      `<button class="filter-btn ${adminFilterSem === "all" ? "active" : ""}" onclick="adminFilterSem='all';renderAdminCourses()">All</button>` +
      sems
        .map(
          (s) =>
            `<button class="filter-btn ${adminFilterSem === s.id ? "active" : ""}" onclick="adminFilterSem=${s.id};renderAdminCourses()">${esc(s.name)}</button>`,
        )
        .join("");
  }

  let html = `
                    <input class="admin-search" placeholder="🔍 Search courses…" value="${adminSearch}" oninput="adminSearch=this.value;renderAdminCourses()"/>
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

  dbPrograms.forEach((prog) => {
    if (adminFilterProg !== "all" && prog.id !== adminFilterProg) return;
    let progHtml = "";
    prog.years.forEach((year) => {
      if (adminFilterYear !== "all" && year.id !== adminFilterYear) return;
      year.sems.forEach((sem) => {
        if (adminFilterSem !== "all" && sem.id !== adminFilterSem) return;
        const filtered = sem.courses.filter(
          (c) =>
            !q ||
            c.name.toLowerCase().includes(q) ||
            c.code.toLowerCase().includes(q),
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
                  ${getLinkBadge(l.type)}<span>${esc(l.label)}</span>${l.note ? `<span style="color:var(--muted)">(${esc(l.note)})</span>` : ""}
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
  const q = adminSearch.toLowerCase();
  let html = `<input class="admin-search" placeholder="🔍 Search extra resources…" value="${adminSearch}" oninput="adminSearch=this.value;renderAdminExtra()"/>
  <button class="btn btn-primary" style="margin-bottom:16px;" onclick="openAddExtraSectionModal()">+ Add Section</button>`;
  dbExtra.forEach((r) => {
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
            ${getLinkBadge(l.type)}<span>${esc(l.label)}</span>${l.note ? `<span style="color:var(--muted)">(${esc(l.note)})</span>` : ""}
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
  const q = adminSearch.toLowerCase();
  try {
    const reports = await sb(
      "reports",
      "GET",
      null,
      null,
      "*&order=created_at.desc",
    );
    let html = `<input class="admin-search" placeholder="🔍 Search reports…" value="${adminSearch}" oninput="adminSearch=this.value;renderAdminReports()"/>`;
    const filtered = reports.filter(
      (r) =>
        !q ||
        r.course_name.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q),
    );
    if (!filtered.length) {
      document.getElementById("adminContent").innerHTML =
        html + '<div class="empty">No reports yet.</div>';
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Course</th><th>Link</th><th>Issue</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    filtered.forEach((r) => {
      html += `<tr><td>${r.course_name}</td><td style="max-width:140px;word-break:break-all;font-size:.75rem;">${r.link_url || "—"}</td><td>${r.description}</td>
        <td><span class="tag ${r.status === "open" ? "tag-open" : "tag-resolved"}">${r.status}</span></td>
        <td class="action-btns">
          <button class="action-btn" onclick="toggleReportStatus(${r.id},'${r.status}')">${r.status === "open" ? "✅ Resolve" : "↩ Reopen"}</button>
          <button class="action-btn del" onclick="confirmAction('Delete this report?',()=>deleteReport(${r.id}))">🗑</button>
        </td></tr>`;
    });
    html += "</tbody></table>";
    document.getElementById("adminContent").innerHTML = html;
  } catch (e) {
    document.getElementById("adminContent").innerHTML =
      `<div class="empty">⚠️ ${e.message}</div>`;
  }
}

// ===================== ADMIN CONTRIBUTIONS =====================
async function renderAdminContributions() {
  document.getElementById("adminContent").innerHTML = getAdminTableSkeleton();
  const q = adminSearch.toLowerCase();
  try {
    const contribs = await sb(
      "contributions",
      "GET",
      null,
      null,
      "*&order=created_at.desc",
    );
    let html = `<input class="admin-search" placeholder="🔍 Search contributions…" value="${adminSearch}" oninput="adminSearch=this.value;renderAdminContributions()"/>`;
    const filtered = contribs.filter(
      (c) =>
        !q ||
        c.course_name.toLowerCase().includes(q) ||
        c.link_url.toLowerCase().includes(q),
    );
    if (!filtered.length) {
      document.getElementById("adminContent").innerHTML =
        html + '<div class="empty">No contributions yet.</div>';
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Course</th><th>Link</th><th>Note</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    filtered.forEach((c) => {
      html += `<tr><td>${c.course_name}</td><td style="max-width:140px;word-break:break-all;font-size:.75rem;">${c.link_url}</td><td>${c.note || "—"}</td>
        <td><span class="tag ${c.status === "pending" ? "tag-open" : "tag-resolved"}">${c.status}</span></td>
        <td class="action-btns">
          <button class="action-btn" onclick="approveContrib(${c.id})">✅ Approve</button>
          <button class="action-btn del" onclick="confirmAction('Reject this contribution?',()=>deleteContrib(${c.id}))">🗑 Reject</button>
        </td></tr>`;
    });
    html += "</tbody></table>";
    document.getElementById("adminContent").innerHTML = html;
  } catch (e) {
    document.getElementById("adminContent").innerHTML =
      `<div class="empty">⚠️ ${e.message}</div>`;
  }
}

// ===================== DB MUTATIONS =====================
async function deleteCourse(id) {
  try {
    await sb(`courses?id=eq.${id}`, "DELETE");
    loadAll();
    showToast("Course deleted.");
  } catch (e) {
    showToast(e.message, true);
  }
}
async function toggleOptional(id, current) {
  try {
    await sb(`courses?id=eq.${id}`, "PATCH", {
      is_optional: !current,
    });
    loadAll();
    renderAdminCourses();
    showToast(current ? "Marked as required." : "Marked as optional.");
  } catch (e) {
    showToast(e.message, true);
  }
}
function confirmDeleteLink(linkId, courseId) {
  // Find the URL of the link being deleted
  let linkUrl = null;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((lk) => {
            if (lk.id === linkId) linkUrl = lk.url;
          }),
        ),
      ),
    ),
  );

  // Find matching links in sibling courses by URL
  const { siblings } = _findSharedCourses(courseId);
  const matchingLinks = [];
  siblings.forEach((sib) =>
    sib.links.forEach((lk) => {
      if (lk.url === linkUrl)
        matchingLinks.push({
          id: lk.id,
          sibName: sib.name,
          prog: sib.prog,
          year: sib.year,
          sem: sib.sem,
        });
    }),
  );

  _pendingLinkOp = {
    linkId,
    matchingLinkIds: matchingLinks.map((m) => m.id),
  };

  if (matchingLinks.length) {
    const list = matchingLinks
      .map(
        (m) =>
          `<li style="margin-bottom:4px;"><strong>${m.sibName}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${m.prog} › ${m.year} › ${m.sem}</span></li>`,
      )
      .join("");
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
  const { linkId, matchingLinkIds } = _pendingLinkOp;
  try {
    await sb(`links?id=eq.${linkId}`, "DELETE");
    if (deleteAll && matchingLinkIds.length)
      await Promise.all(
        matchingLinkIds.map((mid) => sb(`links?id=eq.${mid}`, "DELETE")),
      );
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminCourses();
    showToast(
      deleteAll
        ? `Removed ${matchingLinkIds.length + 1} links!`
        : "Link removed.",
    );
  } catch (e) {
    showToast(e.message, true);
  } finally {
    _pendingLinkOp = null;
  }
}
async function deleteExtraSection(id) {
  try {
    await sb(`extra_sections?id=eq.${id}`, "DELETE");
    loadAll();
    renderAdminExtra();
    showToast("Section deleted.");
  } catch (e) {
    showToast(e.message, true);
  }
}
async function deleteExtraLink(id) {
  try {
    await sb(`extra_links?id=eq.${id}`, "DELETE");
    loadAll();
    renderAdminExtra();
    showToast("Link removed.");
  } catch (e) {
    showToast(e.message, true);
  }
}
async function toggleReportStatus(id, status) {
  try {
    await sb(`reports?id=eq.${id}`, "PATCH", {
      status: status === "open" ? "resolved" : "open",
    });
    renderAdminReports();
    loadReportsBadges();
  } catch (e) {
    showToast(e.message, true);
  }
}
async function deleteReport(id) {
  try {
    await sb(`reports?id=eq.${id}`, "DELETE");
    renderAdminReports();
    loadReportsBadges();
    showToast("Report deleted.");
  } catch (e) {
    showToast(e.message, true);
  }
}
async function approveContrib(id) {
  try {
    await sb(`contributions?id=eq.${id}`, "PATCH", {
      status: "approved",
    });
    renderAdminContributions();
    loadReportsBadges();
    showToast("Marked as approved.");
  } catch (e) {
    showToast(e.message, true);
  }
}
async function deleteContrib(id) {
  try {
    await sb(`contributions?id=eq.${id}`, "DELETE");
    renderAdminContributions();
    loadReportsBadges();
    showToast("Contribution rejected.");
  } catch (e) {
    showToast(e.message, true);
  }
}