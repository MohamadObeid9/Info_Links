
// ===================== CONFIRM MODAL =====================
let _modalCallback = null;
let _lastFocusedElement = null;

function confirmAction(msg, fn) {
  _modalCallback = fn;
  openModal(`<h2>⚠️ Confirm</h2>
    <p style="color:var(--muted);margin-top:8px;font-size:.9rem;">${esc(msg)}</p>
    <div class="modal-actions">
      <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
      <button class="btn" style="background:var(--danger);color:#fff;" onclick="executeModalAction()">Delete</button>
    </div>`);
}

function executeModalAction() {
  if (typeof _modalCallback === "function") _modalCallback();
  closeModal();
  _modalCallback = null;
}

// ── confirmLink: Fixed XSS — URL stored in dataset, not inline onclick ─────
function confirmLink(linkId, rawUrl) {
  // linkId is provided for course links; rawUrl for extra-section links.
  // Fire click tracking if we have a linkId.
  if (linkId) trackLinkClick(linkId);

  // Resolve URL safely from dataset rather than injecting into onclick attr
  let url = rawUrl;
  if (!url && linkId) {
    // Look up url from the in-memory tree
    outer: for (const p of AppState.dbPrograms) {
      for (const y of p.years) {
        for (const s of y.sems) {
          for (const c of s.courses) {
            for (const l of c.links) {
              if (l.id === linkId) { url = l.url; break outer; }
            }
          }
        }
      }
    }
  }
  if (!url) return;
  let parsed;
  try {
    parsed = new URL(url, window.location.origin);
  } catch (err) {
    showToast("Invalid link URL.", true);
    return;
  }
  if (!["http:", "https:"].includes(parsed.protocol)) {
    showToast("Blocked unsafe URL scheme.", true);
    return;
  }

  const box = document.createElement("div");
  box.innerHTML = `<h2>🔗 Open External Link</h2>
  <p style="color:var(--muted);margin-top:8px;font-size:1rem;">You're leaving Info Links to visit:</p>
  <p style="word-break:break-all;font-family:monospace;background:var(--bg3);padding:8px;border-radius:4px;margin:10px 0;font-size:.85rem;">${esc(url)}</p>
  <div class="modal-actions">
    <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
    <button class="btn" id="openLinkBtn">Open Link ↗</button>
  </div>`;

  document.getElementById("modalBox").innerHTML = "";
  document.getElementById("modalBox").appendChild(box);
  document.getElementById("modal").classList.add("open");

  // Attach click handler safely — no inline eval
  document.getElementById("openLinkBtn").addEventListener("click", () => {
    closeModal();
    window.open(parsed.href, "_blank", "noopener,noreferrer");
  });
}

// ===================== MODALS =====================
function openModal(html) {
  const modalEl = document.getElementById("modal");
  const modalBox = document.getElementById("modalBox");
  _lastFocusedElement = document.activeElement;
  modalBox.innerHTML = html;
  modalEl.classList.add("open");
  modalEl.setAttribute("aria-hidden", "false");
  modalBox.setAttribute("role", "dialog");
  modalBox.setAttribute("aria-modal", "true");
  _focusFirstModalElement();
}
function closeModal() {
  const modalEl = document.getElementById("modal");
  modalEl.classList.remove("open");
  modalEl.setAttribute("aria-hidden", "true");
  const modalBox = document.getElementById("modalBox");
  modalBox.removeAttribute("role");
  modalBox.removeAttribute("aria-modal");
  if (_lastFocusedElement && typeof _lastFocusedElement.focus === "function") {
    _lastFocusedElement.focus();
  }
}

function _focusFirstModalElement() {
  const modalBox = document.getElementById("modalBox");
  const firstFocusable = modalBox.querySelector(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  );
  if (firstFocusable && typeof firstFocusable.focus === "function") {
    firstFocusable.focus();
  }
}

function _trapModalFocus(e) {
  const modalEl = document.getElementById("modal");
  if (!modalEl.classList.contains("open")) return;
  if (e.key === "Escape") {
    e.preventDefault();
    closeModal();
    return;
  }
  if (e.key !== "Tab") return;
  const modalBox = document.getElementById("modalBox");
  const focusables = modalBox.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  );
  if (!focusables.length) return;
  const first = focusables[0];
  const last = focusables[focusables.length - 1];
  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault();
    last.focus();
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault();
    first.focus();
  }
}

// ── Content type checkboxes (multi-select) ─────────────────────────────────
function _contentTypeCheckboxes(selectedStr, prefix = "ct") {
  const selected = selectedStr ? String(selectedStr).split(",").map(t => t.trim()) : [];
  const opts = [
    ["td", "✏️ TD"],
    ["cours", "📄 Cours"],
    ["videos", "🎬 Videos"],
    ["sessions", "🎤 Sessions"],
    ["exams", "📝 Exams"],
    ["other", "📦 Other"],
  ];
  return `<div class="content-type-group" id="${prefix}Group">` +
    opts.map(([v, label]) =>
      `<label class="ct-checkbox-label"><input type="checkbox" name="${prefix}" value="${v}" ${selected.includes(v) ? "checked" : ""}/>${label}</label>`
    ).join("") + `</div>`;
}

function _readContentTypeCheckboxes(prefix = "ct") {
  const checks = document.querySelectorAll(`input[name="${prefix}"]:checked`);
  const vals = [...checks].map(c => c.value);
  return vals.length ? vals.join(",") : null;
}

// Keep backward compat for any code using _contentTypeOptions
function _contentTypeOptions(selected) {
  const opts = [
    ["", "— None —"],
    ["td", "✏️ TD"],
    ["cours", "📄 Cours"],
    ["videos", "🎬 Videos"],
    ["sessions", "🎤 Sessions"],
    ["exams", "📝 Exams"],
    ["other", "📦 Other"],
  ];
  return opts
    .map(
      ([v, label]) =>
        `<option value="${v}" ${selected === v ? "selected" : ""}>${label}</option>`,
    )
    .join("");
}

// ── Link type options (reusable) ───────────────────────────────────────────
function _linkTypeOptions(selected) {
  const opts = [
    ["telegram", "Telegram"],
    ["drive", "Google Drive"],
    ["classroom", "Google Classroom"],
    ["other", "Other / External"],
  ];
  return opts
    .map(
      ([v, label]) =>
        `<option value="${v}" ${selected === v ? "selected" : ""}>${label}</option>`,
    )
    .join("");
}

// Add Course
function openAddCourseModal() {
  const progOpts = AppState.dbPrograms
    .map((p) => `<option value="${p.id}">${esc(p.name)}</option>`)
    .join("");
  openModal(`<h2>➕ Add Course</h2>
  <label>Program</label><select id="mProg" onchange="updateYearSemOpts()">${progOpts}</select>
  <label>Year</label><select id="mYear" onchange="updateSemOpts()"></select>
  <label>Semester</label><select id="mSem"></select>
  <label>Course Name</label><input type="text" id="mName" placeholder="e.g. Machine Learning"/>
  <label>Course Code</label><input type="text" id="mCode" placeholder="e.g. ML101"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addCourse()">Add</button></div>`);
  updateYearSemOpts();
}
function updateYearSemOpts() {
  const pi = parseInt(document.getElementById("mProg").value);
  const prog = AppState.dbPrograms.find((p) => p.id === pi);
  document.getElementById("mYear").innerHTML = prog.years
    .map((y) => `<option value="${y.id}">${esc(y.name)}</option>`)
    .join("");
  updateSemOpts();
}
function updateSemOpts() {
  const pi = parseInt(document.getElementById("mProg").value);
  const yi = parseInt(document.getElementById("mYear").value);
  const prog = AppState.dbPrograms.find((p) => p.id === pi);
  const year = prog.years.find((y) => y.id === yi);
  document.getElementById("mSem").innerHTML = year.sems
    .map((s) => `<option value="${s.id}">${esc(s.name)}</option>`)
    .join("");
}
async function addCourse() {
  const semId = parseInt(document.getElementById("mSem").value);
  const name = document.getElementById("mName").value.trim();
  const code = document.getElementById("mCode").value.trim();
  if (!name || !code) {
    showToast("Name and code required.", true);
    return;
  }
  const btn = document.querySelector("#modalBox .btn-primary");
  setBtnLoading(btn, true, "Adding…");
  try {
    await sb("courses", "POST", {
      semester_id: semId,
      name,
      code,
      display_order: 0,
    });
    closeModal();
    _clearCache();
    loadAll();
    showToast("Course added!");
  } catch (e) {
    showToast(e.message, true);
  } finally {
    setBtnLoading(btn, false);
  }
}

// Edit Course
function openEditCourseModal(id) {
  let c, currentProgId, currentYearId, currentSemId;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((co) => {
          if (co.id === id) {
            c = co;
            currentProgId = p.id;
            currentYearId = y.id;
            currentSemId = s.id;
          }
        }),
      ),
    ),
  );
  if (!c) return;

  const progOpts = AppState.dbPrograms
    .map(
      (p) =>
        `<option value="${p.id}" ${p.id === currentProgId ? "selected" : ""}>${esc(p.name)}</option>`,
    )
    .join("");

  const cp = AppState.dbPrograms.find((p) => p.id === currentProgId);
  const yearOpts = cp.years
    .map(
      (y) =>
        `<option value="${y.id}" ${y.id === currentYearId ? "selected" : ""}>${esc(y.name)}</option>`,
    )
    .join("");

  const cy = cp.years.find((y) => y.id === currentYearId);
  const semOpts = cy.sems
    .map(
      (s) =>
        `<option value="${s.id}" ${s.id === currentSemId ? "selected" : ""}>${esc(s.name)}</option>`,
    )
    .join("");

  openModal(`<h2>✏️ Edit Course</h2>
  <label>Course Name</label><input type="text" id="eName" value="${esc(c.name)}"/>
  <label>Course Code</label><input type="text" id="eCode" value="${esc(c.code)}"/>
  <label>Program</label><select id="eProg" onchange="updateEditYearOpts()">${progOpts}</select>
  <label>Year</label><select id="eYear" onchange="updateEditSemOpts()">${yearOpts}</select>
  <label>Semester</label><select id="eSem">${semOpts}</select>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveCourse(${id})">Save</button></div>`);
}
function updateEditYearOpts() {
  const pi = parseInt(document.getElementById("eProg").value);
  const prog = AppState.dbPrograms.find((p) => p.id === pi);
  document.getElementById("eYear").innerHTML = prog.years
    .map((y) => `<option value="${y.id}">${esc(y.name)}</option>`)
    .join("");
  updateEditSemOpts();
}
function updateEditSemOpts() {
  const pi = parseInt(document.getElementById("eProg").value);
  const yi = parseInt(document.getElementById("eYear").value);
  const prog = AppState.dbPrograms.find((p) => p.id === pi);
  const year = prog.years.find((y) => y.id === yi);
  document.getElementById("eSem").innerHTML = year.sems
    .map((s) => `<option value="${s.id}">${esc(s.name)}</option>`)
    .join("");
}
async function saveCourse(id) {
  const name = document.getElementById("eName").value.trim();
  const code = document.getElementById("eCode").value.trim();
  const semId = parseInt(document.getElementById("eSem").value);
  if (!name || !code) {
    showToast("Name and code required.", true);
    return;
  }

  let originalCode = null;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === id) originalCode = c.code;
        }),
      ),
    ),
  );

  const duplicates = [];
  if (originalCode) {
    AppState.dbPrograms.forEach((p) =>
      p.years.forEach((y) =>
        y.sems.forEach((s) =>
          s.courses.forEach((c) => {
            if (c.id !== id && c.code === originalCode)
              duplicates.push({ id: c.id, name: c.name, prog: p.name, year: y.name, sem: s.name });
          }),
        ),
      ),
    );
  }

  AppState._pendingCourseEdit = { id, name, code, semId, duplicateIds: duplicates.map((d) => d.id) };

  if (duplicates.length > 0) {
    const dupList = duplicates
      .map((d) => `<li style="margin-bottom:6px;"><strong>${esc(d.name)}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${esc(d.prog)} › ${esc(d.year)} › ${esc(d.sem)}</span></li>`)
      .join("");
    openModal(`<h2>🔁 Shared Course Detected</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;">
      <strong>${duplicates.length}</strong> other course(s) share the code <strong>${esc(originalCode)}</strong>:
  </p>
  <ul style="font-size:.83rem;margin-bottom:14px;padding-left:18px;list-style:disc;">${dupList}</ul>
  <p style="font-size:.85rem;margin-bottom:6px;">Update <strong>name &amp; code</strong> on all of them too?</p>
  <p style="font-size:.75rem;color:var(--muted);margin-bottom:16px;">The location change only applies to this course.</p>
  <div class="modal-actions">
      <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
      <button class="btn btn-ghost" onclick="applySaveCourse(false)">Just this one</button>
      <button class="btn btn-primary" onclick="applySaveCourse(true)">All ${duplicates.length + 1} occurrences</button>
  </div>`);
  } else {
    await applySaveCourse(false);
  }
}

async function applySaveCourse(updateAll) {
  const { id, name, code, semId, duplicateIds } = AppState._pendingCourseEdit;
  const btn = document.querySelector("#modalBox .btn-primary");
  setBtnLoading(btn, true, "Saving…");
  try {
    await sb(`courses?id=eq.${id}`, "PATCH", { name, code, semester_id: semId });
    if (updateAll && duplicateIds.length) {
      await Promise.all(duplicateIds.map((did) => sb(`courses?id=eq.${did}`, "PATCH", { name, code })));
    }
    closeModal();
    _clearCache();
    loadAll();
    showToast(updateAll ? `Updated all ${duplicateIds.length + 1} occurrences!` : "Course updated!");
  } catch (e) {
    showToast(e.message, true);
  } finally {
    AppState._pendingCourseEdit = null;
    setBtnLoading(btn, false);
  }
}

// Add Link
function openAddLinkModal(courseId) {
  openModal(`<h2>+ Add Link</h2>
  <label>Type</label><select id="lType">${_linkTypeOptions("telegram")}</select>
  <label>URL</label><input type="text" id="lUrl" placeholder="https://…"/>
  <label>Label</label><input type="text" id="lLabel" placeholder="Link 1"/>
  <label>Content Type(s)</label>${_contentTypeCheckboxes("", "lct")}
  <label>Note (optional)</label><input type="text" id="lNote" placeholder="e.g. mail 3adi"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addLink(${courseId})">Add</button></div>`);
}
async function addLink(courseId) {
  const url = document.getElementById("lUrl").value.trim();
  if (!url) { showToast("URL required.", true); return; }
  const type = document.getElementById("lType").value;
  const label = document.getElementById("lLabel").value.trim() || "Link";
  const note = document.getElementById("lNote").value.trim();
  const contentType = _readContentTypeCheckboxes("lct");

  const { code, siblings } = _findSharedCourses(courseId);
  AppState._pendingLinkOp = { courseId, url, type, label, note, contentType, siblingCourseIds: siblings.map((s) => s.id) };

  if (siblings.length) {
    const list = siblings
      .map((s) => `<li style="margin-bottom:4px;"><strong>${esc(s.name)}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${esc(s.prog)} › ${esc(s.year)} › ${esc(s.sem)}</span></li>`)
      .join("");
    openModal(`<h2>🔁 Shared Course</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${siblings.length}</strong> other course(s) share code <strong>${esc(code)}</strong>:</p>
  <ul style="font-size:.83rem;margin-bottom:16px;padding-left:18px;list-style:disc;">${list}</ul>
  <p style="font-size:.85rem;margin-bottom:16px;">Add this link to all of them too?</p>
  <div class="modal-actions">
      <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
      <button class="btn btn-ghost" onclick="applyAddLink(false)">Just this one</button>
      <button class="btn btn-primary" onclick="applyAddLink(true)">All ${siblings.length + 1} courses</button>
  </div>`);
  } else {
    await applyAddLink(false);
  }
}
function _getNextDisplayOrder(courseId) {
  let max = 0;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === courseId)
            c.links.forEach((l) => {
              if ((l.display_order || 0) > max) max = l.display_order || 0;
            });
        }),
      ),
    ),
  );
  return max + 1;
}
async function applyAddLink(addToAll) {
  const { courseId, url, type, label, note, contentType, siblingCourseIds } = AppState._pendingLinkOp;
  try {
    await sb("links", "POST", {
      course_id: courseId, type, url, label, note,
      content_type: contentType,
      display_order: _getNextDisplayOrder(courseId),
    });
    if (addToAll && siblingCourseIds.length)
      await Promise.all(siblingCourseIds.map((sid) =>
        sb("links", "POST", {
          course_id: sid, type, url, label, note,
          content_type: contentType,
          display_order: _getNextDisplayOrder(sid),
        }),
      ));
    closeModal();
    _clearCache();
    loadAll();
    showToast(addToAll ? `Link added to all ${siblingCourseIds.length + 1} courses!` : "Link added!");
  } catch (e) {
    showToast(e.message, true);
  } finally {
    AppState._pendingLinkOp = null;
  }
}

// Edit Link
function openEditLinkModal(linkId, courseId) {
  let l;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((lk) => { if (lk.id === linkId) l = lk; }),
        ),
      ),
    ),
  );
  if (!l) return;
  openModal(`<h2>✏️ Edit Link</h2>
  <label>Type</label>
  <select id="elType">${_linkTypeOptions(l.type)}</select>
  <label>URL</label><input type="text" id="elUrl" value="${esc(l.url)}"/>
  <label>Label</label><input type="text" id="elLabel" value="${esc(l.label)}"/>
  <label>Content Type(s)</label>${_contentTypeCheckboxes(l.content_type || "", "elct")}
  <label>Note (optional)</label><input type="text" id="elNote" value="${esc(l.note || "")}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveLink(${linkId},${courseId})">Save</button></div>`);
}
async function saveLink(linkId, courseId) {
  const url = document.getElementById("elUrl").value.trim();
  if (!url) { showToast("URL required.", true); return; }
  const type = document.getElementById("elType").value;
  const label = document.getElementById("elLabel").value.trim() || "Link";
  const note = document.getElementById("elNote").value.trim();
  const contentType = _readContentTypeCheckboxes("elct");

  let originalUrl = null;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((lk) => { if (lk.id === linkId) originalUrl = lk.url; }),
        ),
      ),
    ),
  );

  const { siblings } = _findSharedCourses(courseId);
  const matchingLinks = [];
  siblings.forEach((sib) =>
    sib.links.forEach((lk) => {
      if (lk.url === originalUrl)
        matchingLinks.push({ id: lk.id, sibName: sib.name, prog: sib.prog, year: sib.year, sem: sib.sem });
    }),
  );

  AppState._pendingLinkOp = { linkId, url, type, label, note, contentType, matchingLinkIds: matchingLinks.map((m) => m.id) };

  if (matchingLinks.length) {
    const list = matchingLinks
      .map((m) => `<li style="margin-bottom:4px;"><strong>${esc(m.sibName)}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${esc(m.prog)} › ${esc(m.year)} › ${esc(m.sem)}</span></li>`)
      .join("");
    openModal(`<h2>🔁 Shared Link</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${matchingLinks.length}</strong> sibling course(s) share the same link:</p>
  <ul style="font-size:.83rem;margin-bottom:16px;padding-left:18px;list-style:disc;">${list}</ul>
  <p style="font-size:.85rem;margin-bottom:16px;">Update all matching links too?</p>
  <div class="modal-actions">
      <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
      <button class="btn btn-ghost" onclick="applySaveLink(false)">Just this one</button>
      <button class="btn btn-primary" onclick="applySaveLink(true)">All ${matchingLinks.length + 1} links</button>
  </div>`);
  } else {
    await applySaveLink(false);
  }
}
async function applySaveLink(updateAll) {
  const { linkId, url, type, label, note, contentType, matchingLinkIds } = AppState._pendingLinkOp;
  try {
    await sb(`links?id=eq.${linkId}`, "PATCH", { type, url, label, note, content_type: contentType });
    if (updateAll && matchingLinkIds.length)
      await Promise.all(matchingLinkIds.map((mid) =>
        sb(`links?id=eq.${mid}`, "PATCH", { type, url, label, note, content_type: contentType }),
      ));
    closeModal();
    _clearCache();
    loadAll();
    showToast(updateAll ? `Updated ${matchingLinkIds.length + 1} links!` : "Link updated!");
  } catch (e) {
    showToast(e.message, true);
  } finally {
    AppState._pendingLinkOp = null;
  }
}

// Extra Sections
function openAddExtraSectionModal() {
  openModal(`<h2>➕ Add Extra Section</h2>
  <label>Icon (emoji)</label><input type="text" id="exIcon" placeholder="📁"/>
  <label>Title</label><input type="text" id="exTitle" placeholder="e.g. Python Resources"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addExtraSection()">Add</button></div>`);
}
async function addExtraSection() {
  const title = document.getElementById("exTitle").value.trim();
  const icon = document.getElementById("exIcon").value.trim() || "📁";
  if (!title) { showToast("Title required.", true); return; }
  try {
    await sb("extra_sections", "POST", { title, icon, display_order: 0 });
    closeModal();
    _clearCache();
    loadAll();
    showToast("Section added!");
  } catch (e) {
    showToast(e.message, true);
  }
}
function openEditExtraSectionModal(id) {
  const r = AppState.dbExtra.find((s) => s.id === id);
  if (!r) return;
  openModal(`<h2>✏️ Edit Section</h2>
  <label>Icon</label><input type="text" id="exIcon" value="${esc(r.icon)}"/>
  <label>Title</label><input type="text" id="exTitle" value="${esc(r.title)}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveExtraSection(${id})">Save</button></div>`);
}
async function saveExtraSection(id) {
  try {
    await sb(`extra_sections?id=eq.${id}`, "PATCH", {
      icon: document.getElementById("exIcon").value.trim() || "📁",
      title: document.getElementById("exTitle").value.trim(),
    });
    closeModal();
    _clearCache();
    loadAll();
    showToast("Section updated!");
  } catch (e) {
    showToast(e.message, true);
  }
}

// Extra Links
function openAddExtraLinkModal(sectionId) {
  openModal(`<h2>+ Add Link</h2>
  <label>Type</label><select id="elxType">${_linkTypeOptions("telegram")}</select>
  <label>URL</label><input type="text" id="elxUrl" placeholder="https://…"/>
  <label>Label</label><input type="text" id="elxLabel" placeholder="Link 1"/>
  <label>Content Type(s)</label>${_contentTypeCheckboxes("", "elxct")}
  <label>Note (optional)</label><input type="text" id="elxNote"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addExtraLink(${sectionId})">Add</button></div>`);
}
async function addExtraLink(sectionId) {
  const url = document.getElementById("elxUrl").value.trim();
  if (!url) { showToast("URL required.", true); return; }
  try {
    await sb("extra_links", "POST", {
      section_id: sectionId,
      type: document.getElementById("elxType").value,
      url,
      label: document.getElementById("elxLabel").value.trim() || "Link",
      note: document.getElementById("elxNote").value.trim(),
      content_type: _readContentTypeCheckboxes("elxct"),
      display_order: 0,
    });
    closeModal();
    _clearCache();
    loadAll();
    showToast("Link added!");
  } catch (e) {
    showToast(e.message, true);
  }
}
function openEditExtraLinkModal(linkId, sectionId) {
  const sec = AppState.dbExtra.find((s) => s.id === sectionId);
  const l = sec && sec.links.find((lk) => lk.id === linkId);
  if (!l) return;
  openModal(`<h2>✏️ Edit Link</h2>
  <label>Type</label>
  <select id="elxType">${_linkTypeOptions(l.type)}</select>
  <label>URL</label><input type="text" id="elxUrl" value="${esc(l.url)}"/>
  <label>Label</label><input type="text" id="elxLabel" value="${esc(l.label)}"/>
  <label>Content Type(s)</label>${_contentTypeCheckboxes(l.content_type || "", "elxct2")}
  <label>Note (optional)</label><input type="text" id="elxNote" value="${esc(l.note || "")}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveExtraLink(${linkId})">Save</button></div>`);
}
async function saveExtraLink(id) {
  const url = document.getElementById("elxUrl").value.trim();
  if (!url) { showToast("URL required.", true); return; }
  try {
    await sb(`extra_links?id=eq.${id}`, "PATCH", {
      type: document.getElementById("elxType").value,
      url,
      label: document.getElementById("elxLabel").value.trim() || "Link",
      note: document.getElementById("elxNote").value.trim(),
      content_type: _readContentTypeCheckboxes("elxct2"),
    });
    closeModal();
    _clearCache();
    loadAll();
    showToast("Link updated!");
  } catch (e) {
    showToast(e.message, true);
  }
}
window.closeModal = closeModal;
window.openAddCourseModal = openAddCourseModal;
document.addEventListener("keydown", _trapModalFocus);
