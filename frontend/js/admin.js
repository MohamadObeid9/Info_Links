import { AppState } from "./state.js";
import { sb, sbAuth, sbLogout } from "./supabase.js";
import { esc, setBtnLoading, getLinkBadge, getContentTypeChips, _findSharedCourses, adminCell, isMobileView } from "./ui.js";
import { getAdminTableSkeleton, getAdminAnalyticsSkeleton } from "./skeleton.js";
import { loadAll, loadReportsBadges } from "./data.js";
import { _clearCache } from "./cache.js";
import { showToast } from "./export.js";
import { renderAdminFeedback } from "./feedback.js";
import { _linkTypeOptions, _contentTypeCheckboxes, _readContentTypeCheckboxes, _getNextDisplayOrder } from "./modals.js";
import { loadStudentDirectory, rememberStudents, senderCell, senderDetail, studentHandleOf, fmtDateTime } from "./students.js";

// ===================== ADMIN AUTH =====================
async function checkLogin() {
  const email = document.getElementById("adminEmail").value.trim();
  const pass = document.getElementById("adminPass").value;
  document.getElementById("loginErr").textContent = "";
  const btn = document.querySelector(".login-wrap .btn-primary");
  setBtnLoading(btn, true, "Logging in…");
  try {
    AppState.sbToken = await sbAuth(email, pass);
    localStorage.setItem("infolinks_token", AppState.sbToken);
    AppState.adminLoggedIn = true;
    document.getElementById("adminPass").value = "";
    window.showView("admin");
  } catch (e) {
    document.getElementById("loginErr").textContent = e.message;
  } finally {
    setBtnLoading(btn, false);
  }
}

async function logout() {
  await sbLogout();
  localStorage.removeItem("infolinks_token");
  AppState.sbToken = null;
  AppState.adminLoggedIn = false;
  window.showView("home");
}

// ===================== ADMIN TABS =====================
const ADMIN_PAGE_SIZE = 10;
const ADMIN_PAGE_FETCH = ADMIN_PAGE_SIZE + 1;
const ANALYTICS_VISITORS_PAGE_SIZE = 12;
const AdminPager = {
  reports: { page: 0, hasNext: false },
  contributions: { page: 0, hasNext: false },
  feedback: { page: 0, hasNext: false },
  students: { page: 0, hasNext: false },
  studentTimeline: { page: 0, hasNext: false },
};

function _setAdminPage(tab, page) {
  if (!AdminPager[tab]) return;
  AdminPager[tab].page = Math.max(0, page);
}

/** Ask for one extra row so a full last page does not look like there is a next one. */
function _pageSlice(rows, pageSize = ADMIN_PAGE_SIZE) {
  const list = Array.isArray(rows) ? rows : [];
  return { items: list.slice(0, pageSize), hasNext: list.length > pageSize };
}

function _pagerButton(label, enabled, onclick) {
  if (!enabled) {
    return `<button type="button" class="action-btn pager-btn" disabled>${label}</button>`;
  }
  return `<button type="button" class="action-btn pager-btn" onclick="${onclick}">${label}</button>`;
}

function _renderAdminPager(tab, rerenderFnName) {
  const pager = AdminPager[tab];
  if (!pager) return "";
  const pageNum = pager.page + 1;
  return `<div class="admin-pager">
    ${_pagerButton("← Prev", pager.page > 0, `adminSetPage('${tab}', -1, '${rerenderFnName}')`)}
    <span>Page ${pageNum}</span>
    ${_pagerButton("Next →", pager.hasNext, `adminSetPage('${tab}', 1, '${rerenderFnName}')`)}
  </div>`;
}

function adminSetPage(tab, delta, rerenderFnName) {
  const pager = AdminPager[tab];
  if (!pager) return;
  if (delta < 0 && pager.page === 0) return;
  if (delta > 0 && !pager.hasNext) return;
  _setAdminPage(tab, pager.page + delta);
  if (typeof window[rerenderFnName] === "function") window[rerenderFnName]();
}

function syncAdminMobileChrome() {
  const select = document.getElementById("adminTabSelect");
  if (select && select.value !== AppState.currentAdminTab) {
    select.value = AppState.currentAdminTab;
  }
  const add = document.getElementById("adminMobileAdd");
  if (add) add.hidden = true;
}

function adminTab(t) {
  AppState.currentAdminTab = t;
  AppState.adminSearch = "";
  AppState.adminFilterProg = "all";
  AppState.adminFilterYear = "all";
  AppState.adminFilterSem = "all";
  AppState.adminStudentId = null;
  _setAdminPage("studentTimeline", 0);
  if (AdminPager[t]) _setAdminPage(t, 0);
  if (t === "feedback" && typeof window.resetAdminFeedbackPage === "function") {
    window.resetAdminFeedbackPage();
  }
  document.querySelectorAll(".admin-tab").forEach((b) => {
    b.classList.toggle("active", b.dataset.adminTab === t);
  });
  syncAdminMobileChrome();
  renderAdminContent();
}

function shortUrl(url) {
  if (!url) return "";
  try {
    const u = new URL(url);
    const host = u.hostname.replace(/^www\./, "");
    const path = u.pathname === "/" ? "" : u.pathname;
    const s = host + path;
    return s.length > 36 ? `${s.slice(0, 34)}…` : s;
  } catch {
    return url.length > 36 ? `${url.slice(0, 34)}…` : url;
  }
}

function _adminLinkRow(l, editOnclick, deleteOnclick) {
  if (!isMobileView()) {
    return `
      <div class="link-chip">
        ${getLinkBadge(l.type)}<span>${esc(l.label)}</span>
        ${getContentTypeChips(l.content_type)}
        ${l.note ? `<span class="admin-muted">(${esc(l.note)})</span>` : ""}
        <button class="action-btn admin-chip-btn" onclick="${editOnclick}">✏️</button>
        <button class="action-btn del admin-chip-btn" onclick="${deleteOnclick}">✕</button>
      </div>`;
  }
  return `
    <div class="link-item admin-link-row">
      <div class="link-item-main">
        ${getLinkBadge(l.type)}
        <span class="link-label">${esc(l.label)}</span>
        ${l.note ? `<span class="link-note">${esc(l.note)}</span>` : ""}
      </div>
      ${getContentTypeChips(l.content_type)}
      <div class="admin-link-actions">
        <button class="action-btn" onclick="${editOnclick}">✏️ Edit</button>
        <button class="action-btn del" onclick="${deleteOnclick}">🗑 Delete</button>
      </div>
    </div>`;
}

function renderAdminContent() {
  loadReportsBadges();
  syncAdminMobileChrome();
  if (AppState.currentAdminTab === "courses") renderAdminCourses();
  else if (AppState.currentAdminTab === "extra") renderAdminExtra();
  else if (AppState.currentAdminTab === "feedback") renderAdminFeedback();
  else if (AppState.currentAdminTab === "reports") renderAdminReports();
  else if (AppState.currentAdminTab === "contributions") renderAdminContributions();
  else if (AppState.currentAdminTab === "students") renderAdminStudents();
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
function resolveLinkInfo(kind, linkId) {
  let info = { label: "Unknown Link", courseName: "Unknown Course" };
  if (kind === "extra_link") {
    AppState.dbExtra.forEach((r) =>
      r.links.forEach((l) => {
        if (l.id == linkId) info = { label: l.label, courseName: r.title };
      }),
    );
    return info;
  }
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((l) => {
            if (l.id == linkId) info = { label: l.label, courseName: c.name };
          }),
        ),
      ),
    ),
  );
  return info;
}

function _num(value) {
  return (Number(value) || 0).toLocaleString();
}

function buildTopLinksList(topLinks, expandKey = null) {
  const rows = (topLinks || [])
    .map((row) => {
      const isExtra = row.extra_link_id != null;
      return {
        kind: isExtra ? "extra_link" : "link",
        id: isExtra ? row.extra_link_id : row.link_id,
        clicks: Number(row.clicks) || 0,
      };
    })
    .filter((row) => row.id != null)
    .sort((a, b) => b.clicks - a.clicks);

  if (!rows.length) {
    return `<div style="color:var(--muted);font-size:0.9rem;">No click data.</div>`;
  }

  const expandable = Boolean(expandKey);
  const expanded = expandable && AppState[expandKey];
  const items = rows
    .slice(0, expanded || !expandable ? 10 : 5)
    .map((row) => {
      const info = resolveLinkInfo(row.kind, row.id);
      return `<li><strong>${_num(row.clicks)}</strong> clicks: ${esc(info.label)} <span style="color:var(--muted);font-size:0.8rem">(${esc(info.courseName)})</span></li>`;
    })
    .join("");

  const expandBtn =
    expandable && rows.length > 5
      ? `<button class="filter-btn" style="margin-top:12px;" onclick="AppState.${expandKey}=!AppState.${expandKey};renderAdminAnalytics()">${expanded ? "Show top 5" : "Show top 10"}</button>`
      : "";

  return `<ul style="list-style:none;padding:0;">${items}</ul>${expandBtn}`;
}

function buildTabbedTopLinksCard(summary) {
  const tab = AppState.analyticsLinksTab === "range" ? "range" : "today";
  const isToday = tab === "today";
  const expandKey = isToday ? "analyticsTopLinksTodayExpanded" : "analyticsTopLinksExpanded";
  const links = isToday ? summary.top_links_today : summary.top_links;

  const tabButtons = ["today", "range"]
    .map((t) => {
      const label = t === "today" ? "Today" : "In range";
      return `<button type="button" class="filter-btn ${tab === t ? "active" : ""}" onclick="AppState.analyticsLinksTab='${t}';renderAdminAnalytics()">${label}</button>`;
    })
    .join("");

  return `<div class="chart-wrap" style="margin-top:20px;">
    <div class="chart-title">🔥 Top clicked links</div>
    <div class="analytics-tabs">${tabButtons}</div>
    ${buildTopLinksList(links, expandKey)}
  </div>`;
}

function buildTopUsersInRangeSection(rows) {
  const list = Array.isArray(rows) ? rows : [];
  if (!list.length) {
    return `<div class="chart-wrap" style="margin-top:20px;">
      <div class="chart-title">🏆 Top students (in range)</div>
      <div style="color:var(--muted);margin-top:16px;font-size:0.9rem;">No clicks in this range yet.</div>
    </div>`;
  }

  const expandKey = "analyticsTopUsersExpanded";
  const expanded = AppState[expandKey];
  const items = list
    .slice(0, expanded ? 10 : 5)
    .map((row) => {
      const id = Number(row.user_id);
      const handle = row.handle || (Number.isFinite(id) ? `#${id}` : "unknown");
      const label = Number.isFinite(id)
        ? `<button class="action-btn" onclick="openAdminStudent(${id})" title="Open student history">${esc(handle)}</button>`
        : esc(handle);
      return `<li style="margin-bottom:8px;"><strong>${_num(row.clicks)}</strong> clicks — ${label}</li>`;
    })
    .join("");

  const expandBtn =
    list.length > 5
      ? `<button class="filter-btn" style="margin-top:12px;" onclick="AppState.${expandKey}=!AppState.${expandKey};renderAdminAnalytics()">${expanded ? "Show top 5" : "Show top 10"}</button>`
      : "";

  return `<div class="chart-wrap" style="margin-top:20px;">
    <div class="chart-title">🏆 Top students (in range)</div>
    <ul style="list-style:none;padding:0;margin-top:16px;">${items}</ul>
    ${expandBtn}
  </div>`;
}

function _visitorsTodayPage(summary) {
  const raw = summary?.visitors_today;
  if (Array.isArray(raw)) {
    return {
      visitors: raw,
      hasMore: Boolean(summary.visitors_has_more),
      total: Number(summary.visitors_today_total) || Number(summary.active_today) || raw.length,
    };
  }
  return {
    visitors: Array.isArray(raw?.visitors) ? raw.visitors : [],
    hasMore: Boolean(raw?.has_more),
    total: Number(summary.active_today) || 0,
  };
}

function buildVisitorChipsSection(summary) {
  const { visitors, hasMore, total } = _visitorsTodayPage(summary);
  const sort = AppState.analyticsVisitorsSort === "name" ? "name" : "clicks";
  const sortButtons = ["clicks", "name"]
    .map((s) => {
      const label = s === "clicks" ? "Most clicks" : "Name";
      return `<button type="button" class="filter-btn ${sort === s ? "active" : ""}" onclick="AppState.analyticsVisitorsSort='${s}';AppState.analyticsVisitorsOffset=0;renderAdminAnalytics()">${label}</button>`;
    })
    .join("");

  if (!visitors.length) {
    return `<div class="chart-wrap" style="margin-top:20px;">
      <div class="chart-title">👀 Who visited today</div>
      <div class="analytics-tabs">${sortButtons}</div>
      <div style="color:var(--muted);margin-top:16px;font-size:0.9rem;">Nobody has visited yet today.</div>
    </div>`;
  }

  const chips = visitors
    .map((row) => {
      const id = Number(row.user_id);
      const handle = row.handle || (Number.isFinite(id) ? `#${id}` : "unknown");
      const clicks = Number(row.clicks) || 0;
      const badge = clicks > 0 ? `<span class="badge-count">${_num(clicks)}</span>` : "";
      const onclick = Number.isFinite(id) ? `openAdminStudent(${id})` : "";
      return `<button type="button" class="visitor-chip" ${onclick ? `onclick="${onclick}"` : ""} title="Open student history">
        <span class="visitor-chip-handle">${esc(handle)}</span>${badge}
      </button>`;
    })
    .join("");

  const offset = Number(AppState.analyticsVisitorsOffset) || 0;
  const pageSize = ANALYTICS_VISITORS_PAGE_SIZE;
  const pageNum = Math.floor(offset / pageSize) + 1;
  const hasPrev = offset > 0;
  const totalHint = Number.isFinite(total) && total > 0
    ? `<span style="font-size:0.8rem;color:var(--muted);margin-left:8px;">${_num(total)} total</span>`
    : "";

  const pager = `<div class="admin-pager">
    ${_pagerButton("← Prev", hasPrev, `AppState.analyticsVisitorsOffset=Math.max(0,${offset}-${pageSize});renderAdminAnalytics()`)}
    <span>Page ${pageNum}</span>
    ${_pagerButton("Next →", hasMore, `AppState.analyticsVisitorsOffset=${offset + pageSize};renderAdminAnalytics()`)}
  </div>`;

  return `<div class="chart-wrap" style="margin-top:20px;">
    <div class="chart-title">👀 Who visited today${totalHint}</div>
    <div class="analytics-tabs">${sortButtons}</div>
    <div class="visitor-chips">${chips}</div>
    ${pager}
  </div>`;
}

function _deviceTodayParts(devices) {
  if (!devices || typeof devices !== "object") return { val: "0", sub: "" };
  const phone = Number(devices.phone) || 0;
  const laptop = Number(devices.laptop) || 0;
  if (phone && laptop) return { val: `${_num(phone)}/${_num(laptop)}`, sub: "phone / laptop" };
  if (phone) return { val: _num(phone), sub: "phone" };
  if (laptop) return { val: _num(laptop), sub: "laptop" };
  return { val: "0", sub: "" };
}

function _formatGain(value) {
  const n = Number(value) || 0;
  return `+${_num(n)}`;
}

/** Pad the server's sparse per-day series so the chart always spans the range. */
function _dailyUniqueSeries(daily, rangeDays) {
  const byDay = new Map();
  (daily || []).forEach((d) => {
    const key = String(d?.day || "").slice(0, 10);
    if (key) byDay.set(key, Number(d.users) || 0);
  });
  const now = Date.now();
  const days = [];
  for (let i = rangeDays - 1; i >= 0; i--) {
    const key = new Date(now - i * 86400000).toISOString().slice(0, 10);
    days.push({ date: key, count: byDay.get(key) || 0 });
  }
  return days;
}

/** Pad cumulative roster totals per day for the growth chart. */
function _dailyRosterSeries(daily, rangeDays) {
  const byDay = new Map();
  (daily || []).forEach((d) => {
    const key = String(d?.day || "").slice(0, 10);
    if (key) byDay.set(key, Number(d.total) || 0);
  });
  const now = Date.now();
  const days = [];
  let lastKnown = 0;
  for (let i = rangeDays - 1; i >= 0; i--) {
    const key = new Date(now - i * 86400000).toISOString().slice(0, 10);
    if (byDay.has(key)) lastKnown = byDay.get(key);
    days.push({ date: key, count: lastKnown });
  }
  return days;
}

function _buildBarChart(days, range, todayStr) {
  const maxCount = Math.max(...days.map((d) => d.count), 1);
  const labelStep = range === "7" ? 1 : range === "30" ? 5 : 15;

  function fmtDay(dateStr) {
    const d = new Date(dateStr + "T00:00:00");
    return d.toLocaleDateString("en", { month: "short", day: "numeric" });
  }

  return days
    .map((d, i) => {
      const pct = Math.round((d.count / maxCount) * 100);
      const showLabel = i % labelStep === 0 || i === days.length - 1;
      return `<div class="bar-col"><div class="bar-val" style="visibility:${d.count > 0 ? "visible" : "hidden"}">${d.count || ""}</div><div class="bar-fill" style="height:${Math.max(pct, d.count > 0 ? 4 : 0)}%;background:${d.date === todayStr ? "var(--accent2)" : "var(--accent)"}"></div><div class="bar-label">${showLabel ? fmtDay(d.date) : ""}</div></div>`;
    })
    .join("");
}

async function renderAdminAnalytics() {
  document.getElementById("adminContent").innerHTML = getAdminAnalyticsSkeleton();
  const range = ["7", "30", "90"].includes(String(AppState.analyticsRange))
    ? String(AppState.analyticsRange)
    : "30";
  const chartSeries = AppState.analyticsChartSeries === "roster" ? "roster" : "visitors";
  const visitorsSort = AppState.analyticsVisitorsSort === "name" ? "name" : "clicks";
  const visitorsOffset = Math.max(0, Number(AppState.analyticsVisitorsOffset) || 0);
  try {
    const query = new URLSearchParams({
      range,
      visitors_limit: String(ANALYTICS_VISITORS_PAGE_SIZE),
      visitors_offset: String(visitorsOffset),
      visitors_sort: visitorsSort,
    });
    const summary = (await sb(`analytics/summary?${query}`, "GET")) || {};

    if (visitorsOffset > 0 && !_visitorsTodayPage(summary).visitors.length) {
      AppState.analyticsVisitorsOffset = Math.max(0, visitorsOffset - ANALYTICS_VISITORS_PAGE_SIZE);
      renderAdminAnalytics();
      return;
    }

    const rangeDays = parseInt(range, 10);
    const todayStr = new Date().toISOString().slice(0, 10);
    const chartDays =
      chartSeries === "roster"
        ? _dailyRosterSeries(summary.daily_roster, rangeDays)
        : _dailyUniqueSeries(summary.daily_unique_visits, rangeDays);
    const barsHtml = _buildBarChart(chartDays, range, todayStr);

    const rangeButtons = ["7", "30", "90"]
      .map((r) => `<button type="button" class="filter-btn ${range === r ? "active" : ""}" onclick="AppState.analyticsRange='${r}';renderAdminAnalytics()">${r} days</button>`)
      .join("");

    const seriesButtons = ["visitors", "roster"]
      .map((s) => {
        const label = s === "visitors" ? "Unique visitors" : "Registered students";
        return `<button type="button" class="filter-btn ${chartSeries === s ? "active" : ""}" onclick="AppState.analyticsChartSeries='${s}';renderAdminAnalytics()">${label}</button>`;
      })
      .join("");

    const chartTitle =
      chartSeries === "roster"
        ? `Registered students over time — <span style="color:var(--accent2);">■</span> today`
        : `Unique students per day — <span style="color:var(--accent2);">■</span> today`;

    const gained7 = Number(summary.students_gained_7d) || 0;
    const deviceToday = _deviceTodayParts(summary.devices_today);
    const deltaRow = `<div class="analytics-deltas">
      <span class="stat-delta">${_formatGain(gained7)} this week</span>
      <span class="stat-delta-sep">·</span>
      <span class="stat-delta">${_formatGain(summary.students_gained_30d)} last 30 days</span>
      <span class="stat-delta-sep">·</span>
      <span class="stat-delta">${_formatGain(summary.students_gained_90d)} last 90 days</span>
    </div>`;

    document.getElementById("adminContent").innerHTML = `
      <div class="stat-grid analytics-kpis">
          <div class="stat-card">
            <div class="stat-val">${_num(summary.total_students)}</div>
            <div class="stat-mid"><span class="stat-delta">${_formatGain(gained7)} this week</span></div>
            <div class="stat-label">Registered students</div>
          </div>
          <div class="stat-card">
            <div class="stat-val">${_num(summary.active_today)}</div>
            <div class="stat-mid"></div>
            <div class="stat-label">Active today</div>
          </div>
          <div class="stat-card">
            <div class="stat-val">${_num(summary.clicks_today)}</div>
            <div class="stat-mid"></div>
            <div class="stat-label">Clicks today</div>
          </div>
          <div class="stat-card">
            <div class="stat-val">${deviceToday.val}</div>
            <div class="stat-mid">${deviceToday.sub ? `<span class="stat-sub">${deviceToday.sub}</span>` : ""}</div>
            <div class="stat-label">Device today</div>
          </div>
      </div>
      <div class="chart-wrap">
          <div class="chart-title">Growth</div>
          ${deltaRow}
          <div class="analytics-range">${rangeButtons}</div>
          <div class="analytics-chart-series">${seriesButtons}</div>
          <div class="chart-title" style="margin-top:0;margin-bottom:12px;font-size:0.88rem;font-weight:600;">${chartTitle}</div>
          <div class="bar-chart-scroll"><div class="bar-chart">${barsHtml}</div></div>
      </div>
      ${buildVisitorChipsSection(summary)}
      ${buildTabbedTopLinksCard(summary)}
      ${buildTopUsersInRangeSection(summary.top_users)}
      <p style="font-size:.78rem;color:var(--muted);margin-top:8px;">Unique students, not raw hits — each student counts once per day however many times they open the site.</p>`;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ Could not load analytics: ${esc(e.message)}</div>`;
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
          progHtml += `<div class="admin-entity-card">
            <div class="admin-entity-head">
              <button type="button" class="admin-entity-toggle">
                <span class="admin-entity-title">
                  <strong>${esc(c.name)}</strong>
                  <span class="course-code">${esc(c.code)}</span>
                  ${c.is_optional ? '<span class="optional-tag">OPTIONAL</span>' : ""}
                </span>
                <span class="admin-entity-hint">${c.links.length} link${c.links.length === 1 ? "" : "s"}</span>
                <span class="course-chev" aria-hidden="true">›</span>
              </button>
              <div class="action-btns">
                <button class="action-btn" onclick="toggleOptional(${c.id},${c.is_optional})">${c.is_optional ? "✅ Optional" : "⬜ Optional"}</button>
                <button class="action-btn" onclick="openEditCourseModal(${c.id})">✏️ Edit</button>
                <button class="action-btn" onclick="openAddLinkModal(${c.id})">+ Link</button>
                <button class="action-btn del" onclick="confirmAction('Delete this course and all its links?',()=>deleteCourse(${c.id}))">🗑 Delete</button>
              </div>
            </div>
            <div class="admin-link-list">
              ${c.links.length
              ? c.links
                .map((l) => _adminLinkRow(
                  l,
                  `openEditLinkModal(${l.id},${c.id})`,
                  `confirmDeleteLink(${l.id},${c.id})`,
                ))
                .join("")
              : '<span class="admin-muted">No links</span>'
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
  <button class="btn btn-primary admin-desktop-add" style="margin-bottom:16px;" onclick="openAddExtraSectionModal()">+ Add Section</button>`;
  AppState.dbExtra.forEach((r) => {
    if (q && !r.title.toLowerCase().includes(q)) return;
    html += `<div class="admin-entity-card">
      <div class="admin-entity-head">
        <button type="button" class="admin-entity-toggle">
          <span class="admin-entity-title">${r.icon} ${esc(r.title)}</span>
          <span class="admin-entity-hint">${r.links.length} link${r.links.length === 1 ? "" : "s"}</span>
          <span class="course-chev" aria-hidden="true">›</span>
        </button>
        <div class="action-btns">
          <button class="action-btn" onclick="openEditExtraSectionModal(${r.id})">✏️ Edit</button>
          <button class="action-btn" onclick="openAddExtraLinkModal(${r.id})">+ Link</button>
          <button class="action-btn del" onclick="confirmAction('Delete this section and all its links?',()=>deleteExtraSection(${r.id}))">🗑 Delete</button>
        </div>
      </div>
      <div class="admin-link-list">
        ${r.links.length
        ? r.links
          .map((l) => _adminLinkRow(
            l,
            `openEditExtraLinkModal(${l.id},${r.id})`,
            `confirmAction('Remove this link?',()=>deleteExtraLink(${l.id},${r.id}))`,
          ))
          .join("")
        : '<span class="admin-muted">No links</span>'
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
    const [fetchedReports] = await Promise.all([
      sb(`reports?limit=${ADMIN_PAGE_FETCH}&offset=${offset}&q=${encodeURIComponent(q)}`, "GET").then((r) => r || []),
      loadStudentDirectory(),
    ]);
    const reportsPage = _pageSlice(fetchedReports);
    const reports = reportsPage.items;
    if (page > 0 && reports.length === 0) {
      _setAdminPage("reports", page - 1);
      renderAdminReports();
      return;
    }
    AdminPager.reports.hasNext = reportsPage.hasNext;
    let html = `<input class="admin-search" placeholder="🔍 Search reports…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;_setAdminPage('reports',0);renderAdminReports()"/>`;
    if (!reports.length) {
      const emptyMsg = q ? `No report matching "${esc(q)}" found.` : "No reports yet.";
      document.getElementById("adminContent").innerHTML = html + `<div class="empty">${emptyMsg}</div>`;
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Sender</th><th>Course</th><th>Link</th><th>Issue</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    reports.forEach((r) => {
      const issue = (r.description || "").trim() || "No description";
      html += `<tr class="admin-row">
        ${senderDetail(r.user_id)}
        ${adminCell("admin-pri", "Course", esc(r.course_name))}
        ${adminCell(r.link_url ? "admin-detail" : "admin-detail admin-empty", "Link", `<span class="admin-url">${esc(r.link_url || "—")}</span>`)}
        ${adminCell("admin-sec", "Issue", esc(issue))}
        ${adminCell("admin-meta", "Status", `<span class="tag ${r.status === "open" ? "tag-open" : "tag-resolved"}">${r.status}</span>`)}
        ${adminCell("admin-actions action-btns", "Actions", `
          <button class="action-btn" onclick="toggleReportStatus(${r.id},'${r.status}')">${r.status === "open" ? "✅ Resolve" : "↩ Reopen"}</button>
          <button class="action-btn del" onclick="confirmAction('Delete this report?',()=>deleteReport(${r.id}))">🗑 Delete</button>
        `)}
      </tr>`;
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
    const [fetchedContribs] = await Promise.all([
      sb(`contributions?limit=${ADMIN_PAGE_FETCH}&offset=${offset}&q=${encodeURIComponent(q)}`, "GET").then((r) => r || []),
      loadStudentDirectory(),
    ]);
    const contribsPage = _pageSlice(fetchedContribs);
    const contribs = contribsPage.items;
    if (page > 0 && contribs.length === 0) {
      _setAdminPage("contributions", page - 1);
      renderAdminContributions();
      return;
    }
    AdminPager.contributions.hasNext = contribsPage.hasNext;
    let html = `<input class="admin-search" placeholder="🔍 Search contributions…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;_setAdminPage('contributions',0);renderAdminContributions()"/>`;
    if (!contribs.length) {
      const emptyMsg = q ? `No contribution matching "${esc(q)}" found.` : "No contributions yet.";
      document.getElementById("adminContent").innerHTML = html + `<div class="empty">${emptyMsg}</div>`;
      return;
    }
    html += `<table class="admin-table"><thead><tr><th>Sender</th><th>Course</th><th>Link</th><th>Note</th><th>Status</th><th>Actions</th></tr></thead><tbody>`;
    contribs.forEach((c) => {
      const statusTag = c.status === "pending"
        ? "tag-open"
        : c.status === "rejected"
          ? "tag-rejected"
          : "tag-resolved";
      const note = (c.note || "").trim();
      const preview = note || shortUrl(c.link_url) || "No note";
      html += `<tr class="admin-row">
        ${senderDetail(c.user_id)}
        ${adminCell("admin-pri", "Course", esc(c.course_name))}
        ${adminCell(c.link_url ? "admin-detail" : "admin-detail admin-empty", "Link", `<span class="admin-url">${esc(c.link_url || "—")}</span>`)}
        ${adminCell("admin-sec", "Note", esc(preview))}
        ${adminCell("admin-meta", "Status", `<span class="tag ${statusTag}">${esc(c.status || "pending")}</span>`)}
        ${adminCell("admin-actions action-btns", "Actions", _contributionActions(c))}
      </tr>`;
    });
    html += "</tbody></table>";
    html += _renderAdminPager("contributions", "renderAdminContributions");
    document.getElementById("adminContent").innerHTML = html;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ ${e.message}</div>`;
  }
}

function _contributionActions(c) {
  if (c.status === "rejected") {
    return `<button class="action-btn" onclick="setContributionStatus(${c.id},'pending','Contribution reopened.')">↩ Reopen</button>
      <button class="action-btn del" onclick="confirmAction('Delete this contribution permanently?',()=>deleteContrib(${c.id}))">🗑</button>`;
  }
  if (c.status === "approved") {
    return `<button class="action-btn" disabled>Approved</button>`;
  }
  return `<button class="action-btn" style="color:var(--success); border-color:var(--success)" onclick="openAutoApproveContribModal(${esc(JSON.stringify(c))})">✅ Approve & Add</button>
    <button class="action-btn del" onclick="confirmAction('Reject this contribution? It stays in the list as rejected.',()=>rejectContrib(${c.id}),'Reject')">✕ Reject</button>`;
}

async function rejectContrib(id) {
  await setContributionStatus(id, "rejected", "Contribution rejected.");
}

async function setContributionStatus(id, status, toast) {
  try {
    await sb(`contributions?id=eq.${id}`, "PATCH", { status });
    renderAdminContributions();
    loadReportsBadges();
    if (toast) showToast(toast);
  } catch (e) { showToast(e.message, true); }
}

// ===================== ADMIN STUDENTS =====================
const TIMELINE_META = {
  visit: { icon: "👣", label: "Visit" },
  link_click: { icon: "🔗", label: "Link opened" },
  report: { icon: "🚨", label: "Report" },
  contribution: { icon: "➕", label: "Contribution" },
  feedback: { icon: "⭐", label: "Feedback" },
  favorite_added: { icon: "★", label: "Favorite added" },
  favorite_removed: { icon: "☆", label: "Favorite removed" },
};

function openAdminStudent(id) {
  AppState.currentAdminTab = "students";
  AppState.adminStudentId = Number(id);
  AppState.adminSearch = "";
  _setAdminPage("studentTimeline", 0);
  document.querySelectorAll(".admin-tab").forEach((b) => {
    b.classList.toggle("active", b.dataset.adminTab === "students");
  });
  renderAdminStudents();
}

function closeAdminStudent() {
  AppState.adminStudentId = null;
  _setAdminPage("studentTimeline", 0);
  renderAdminStudents();
}

async function renderAdminStudents() {
  if (AppState.adminStudentId) {
    renderAdminStudentDetail();
    return;
  }
  document.getElementById("adminContent").innerHTML = getAdminTableSkeleton();
  const q = AppState.adminSearch.trim();
  const page = AdminPager.students.page;
  const offset = page * ADMIN_PAGE_SIZE;
  try {
    const studentsPage = _pageSlice(
      (await sb(`users?limit=${ADMIN_PAGE_FETCH}&offset=${offset}&q=${encodeURIComponent(q)}`, "GET")) || [],
    );
    const students = studentsPage.items;
    if (page > 0 && students.length === 0) {
      _setAdminPage("students", page - 1);
      renderAdminStudents();
      return;
    }
    AdminPager.students.hasNext = studentsPage.hasNext;
    rememberStudents(students);

    let html = `<input class="admin-search" placeholder="🔍 Search students…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;_setAdminPage('students',0);renderAdminStudents()"/>`;
    if (!students.length) {
      const emptyMsg = q ? `No student matching "${esc(q)}" found.` : "No students yet.";
      document.getElementById("adminContent").innerHTML = html + `<div class="empty">${emptyMsg}</div>`;
      return;
    }

    html += `<table class="admin-table"><thead><tr><th>Student</th><th>First seen</th><th>Last seen</th><th>Visits</th><th>Clicks</th><th>Actions</th></tr></thead><tbody>`;
    students.forEach((u) => {
      html += `<tr class="admin-row" data-student-id="${Number(u.id)}">
        ${adminCell("admin-pri", "Student", `<strong>${esc(studentHandleOf(u))}</strong>`)}
        ${adminCell("admin-sec", "Last seen", esc(fmtDateTime(u.last_seen_at)))}
        ${adminCell("admin-detail", "First seen", esc(fmtDateTime(u.created_at)))}
        ${adminCell("admin-meta", "Visits", String(_num(u.visit_count)))}
        ${adminCell("admin-detail", "Clicks", String(_num(u.click_count)))}
        ${adminCell("admin-actions action-btns", "Actions", `<button class="action-btn" onclick="openAdminStudent(${Number(u.id)})">👤 History</button>`)}
      </tr>`;
    });
    html += "</tbody></table>";
    html += _renderAdminPager("students", "renderAdminStudents");
    document.getElementById("adminContent").innerHTML = html;
    _refocusSearch();
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `<div class="empty">⚠️ ${esc(e.message)}</div>`;
  }
}

function _favoriteCourseNames(ids) {
  const names = (ids || [])
    .map((id) => AppState.courseById.get(Number(id))?.name || `#${id}`)
    .map((name) => esc(String(name)));
  return names.length ? names.join(", ") : "—";
}

async function renderAdminStudentDetail() {
  const id = AppState.adminStudentId;
  document.getElementById("adminContent").innerHTML = getAdminTableSkeleton();
  const page = AdminPager.studentTimeline.page;
  const offset = page * ADMIN_PAGE_SIZE;
  try {
    const data =
      (await sb(`users/${id}?limit=${ADMIN_PAGE_FETCH}&offset=${offset}`, "GET")) || {};
    const user = data.user || {};
    const timelinePage = _pageSlice(data.timeline);
    const timeline = timelinePage.items;
    if (page > 0 && timeline.length === 0) {
      _setAdminPage("studentTimeline", page - 1);
      renderAdminStudentDetail();
      return;
    }
    AdminPager.studentTimeline.hasNext = timelinePage.hasNext;
    rememberStudents([user]);

    const rows = timeline.length
      ? timeline
        .map((item) => {
          const meta = TIMELINE_META[item.type] || { icon: "•", label: item.type || "Activity" };
          return `<tr class="admin-row">
              ${adminCell("admin-pri", "Type", `${meta.icon} ${esc(meta.label)}`)}
              ${adminCell("admin-sec", "Details", esc(item.summary || "—"))}
              ${adminCell("admin-meta", "When", esc(fmtDateTime(item.at)))}
            </tr>`;
        })
        .join("")
      : `<tr class="admin-table-empty"><td colspan="3" style="color:var(--muted);">No activity on this page.</td></tr>`;

    document.getElementById("adminContent").innerHTML = `
      <button class="action-btn" style="margin-bottom:16px;" onclick="closeAdminStudent()">← All students</button>
      <div class="stat-grid">
        <div class="stat-card"><div class="stat-val" style="font-size:1.2rem;word-break:break-all;">${esc(studentHandleOf(user))}</div><div class="stat-label">Student</div></div>
        <div class="stat-card"><div class="stat-val" style="font-size:1rem;">${esc(fmtDateTime(user.created_at))}</div><div class="stat-label">Signed up</div></div>
        <div class="stat-card"><div class="stat-val" style="font-size:1rem;">${esc(fmtDateTime(user.last_seen_at))}</div><div class="stat-label">Last seen</div></div>
      </div>
      <div class="chart-wrap" style="margin-bottom:20px;">
        <div class="chart-title">⭐ Current favorites</div>
        <div style="font-size:.85rem;color:var(--muted);">${_favoriteCourseNames(user.favorite_course_ids)}</div>
      </div>
      <div class="chart-title">🕒 Activity history</div>
      <table class="admin-table"><thead><tr><th>Type</th><th>When</th><th>Details</th></tr></thead><tbody>${rows}</tbody></table>
      ${_renderAdminPager("studentTimeline", "renderAdminStudentDetail")}`;
  } catch (e) {
    document.getElementById("adminContent").innerHTML = `
      <button class="action-btn" style="margin-bottom:16px;" onclick="closeAdminStudent()">← All students</button>
      <div class="empty">⚠️ ${esc(e.message)}</div>`;
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

  window.openModal(`<h2>✅ Approve Contribution</h2>
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
    window.openModal(`<h2>🔁 Shared Course</h2>
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
      window.closeModal(); _clearCache(); loadAll(); renderAdminContributions();
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
    window.closeModal(); _clearCache(); loadAll(); renderAdminContributions();
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
    window.openModal(`<h2>🗑 Delete Link</h2>
    <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${matchingLinks.length}</strong> sibling course(s) have the same link:</p>
    <ul style="font-size:.83rem;margin-bottom:16px;padding-left:18px;list-style:disc;">${list}</ul>
    <p style="font-size:.85rem;margin-bottom:16px;">Remove from all of them too?</p>
    <div class="modal-actions">
        <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
        <button class="btn btn-ghost" onclick="applyDeleteLink(false)">Just this one</button>
        <button class="btn" style="background:var(--danger);color:#fff;" onclick="applyDeleteLink(true)">All ${matchingLinks.length + 1} links</button>
    </div>`);
  } else {
    window.openModal(`<h2>⚠️ Confirm</h2>
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
    window.closeModal();
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
    showToast("Contribution deleted.");
  } catch (e) { showToast(e.message, true); }
}

function bindAdminMobile() {
  const select = document.getElementById("adminTabSelect");
  if (select && !select.dataset.bound) {
    select.addEventListener("change", () => adminTab(select.value));
    select.dataset.bound = "1";
  }
  const root = document.getElementById("view-admin");
  if (root && !root.dataset.mobileBound) {
    root.addEventListener("click", (e) => {
      if (!isMobileView()) return;
      if (e.target.closest(".action-btn, a, select, input, .btn")) return;

      const toggle = e.target.closest(".admin-entity-toggle");
      if (toggle) {
        e.preventDefault();
        const card = toggle.closest(".admin-entity-card");
        const open = card.classList.contains("open");
        document.querySelectorAll(".admin-entity-card.open").forEach((el) => el.classList.remove("open"));
        if (!open) card.classList.add("open");
        return;
      }

      const studentRow = e.target.closest("tr[data-student-id]");
      if (studentRow) {
        openAdminStudent(studentRow.dataset.studentId);
        return;
      }

      const row = e.target.closest(".admin-table tr.admin-row");
      if (row && !row.classList.contains("admin-table-empty")) {
        const open = row.classList.contains("open");
        document.querySelectorAll(".admin-table tr.open").forEach((el) => el.classList.remove("open"));
        if (!open) row.classList.add("open");
      }
    });
    root.dataset.mobileBound = "1";
  }
}

bindAdminMobile();

Object.assign(window, {
  checkLogin,
  logout,
  adminTab,
  renderAdminContent,
  adminSetPage,
  _setAdminPage,
  renderAdminReports,
  renderAdminContributions,
  renderAdminAnalytics,
  renderAdminCourses,
  renderAdminExtra,
  toggleOptional,
  deleteCourse,
  confirmDeleteLink,
  applyDeleteLink,
  deleteExtraSection,
  deleteExtraLink,
  toggleReportStatus,
  deleteReport,
  openAutoApproveContribModal,
  applyAutoApproveContrib,
  _applyAutoApproveWithSiblings,
  rejectContrib,
  setContributionStatus,
  deleteContrib,
  openAdminStudent,
  closeAdminStudent,
  renderAdminStudents,
  renderAdminStudentDetail,
  ADMIN_PAGE_SIZE,
});

export {
  renderAdminContent,
  renderAdminAnalytics,
  renderAdminCourses,
  renderAdminExtra,
  renderAdminReports,
  renderAdminContributions,
};
