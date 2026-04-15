// ===================== CONFIRM MODAL =====================
  function confirmAction(msg, fn) {
    openModal(`<h2>⚠️ Confirm</h2>
    <p style="color:var(--muted);margin-top:8px;font-size:.9rem;">${msg}</p>
    <div class="modal-actions">
      <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
      <button class="btn" style="background:var(--danger);color:#fff;" onclick="closeModal();(${fn})()">Delete</button>
    </div>`);
  }

// ===================== MODALS =====================
function openModal(html) {
  document.getElementById("modalBox").innerHTML = html;
  document.getElementById("modal").classList.add("open");
}
function closeModal() {
  document.getElementById("modal").classList.remove("open");
}

// Add Course
function openAddCourseModal() {
  const progOpts = dbPrograms
    .map((p) => `<option value="${p.id}">${p.name}</option>`)
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
  const prog = dbPrograms.find((p) => p.id === pi);
  document.getElementById("mYear").innerHTML = prog.years
    .map((y) => `<option value="${y.id}">${y.name}</option>`)
    .join("");
  updateSemOpts();
}
function updateSemOpts() {
  const pi = parseInt(document.getElementById("mProg").value);
  const yi = parseInt(document.getElementById("mYear").value);
  const prog = dbPrograms.find((p) => p.id === pi);
  const year = prog.years.find((y) => y.id === yi);
  document.getElementById("mSem").innerHTML = year.sems
    .map((s) => `<option value="${s.id}">${s.name}</option>`)
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
  try {
    await sb("courses", "POST", {
      semester_id: semId,
      name,
      code,
      display_order: 0,
    });
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminCourses();
    showToast("Course added!");
  } catch (e) {
    showToast(e.message, true);
  }
}

// Edit Course
function openEditCourseModal(id) {
  let c, currentProgId, currentYearId, currentSemId;
  dbPrograms.forEach((p) =>
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

  const progOpts = dbPrograms
    .map(
      (p) =>
        `<option value="${p.id}" ${p.id === currentProgId ? "selected" : ""}>${p.name}</option>`,
    )
    .join("");

  const currentProg = dbPrograms.find((p) => p.id === currentProgId);
  const yearOpts = currentProg.years
    .map(
      (y) =>
        `<option value="${y.id}" ${y.id === currentYearId ? "selected" : ""}>${y.name}</option>`,
    )
    .join("");

  const currentYear = currentProg.years.find((y) => y.id === currentYearId);
  const semOpts = currentYear.sems
    .map(
      (s) =>
        `<option value="${s.id}" ${s.id === currentSemId ? "selected" : ""}>${s.name}</option>`,
    )
    .join("");

  openModal(`<h2>✏️ Edit Course</h2>
  <label>Course Name</label><input type="text" id="eName" value="${c.name}"/>
  <label>Course Code</label><input type="text" id="eCode" value="${c.code}"/>
  <label>Program</label><select id="eProg" onchange="updateEditYearOpts()">${progOpts}</select>
  <label>Year</label><select id="eYear" onchange="updateEditSemOpts()">${yearOpts}</select>
  <label>Semester</label><select id="eSem">${semOpts}</select>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveCourse(${id})">Save</button></div>`);
}
function updateEditYearOpts() {
  const pi = parseInt(document.getElementById("eProg").value);
  const prog = dbPrograms.find((p) => p.id === pi);
  document.getElementById("eYear").innerHTML = prog.years
    .map((y) => `<option value="${y.id}">${y.name}</option>`)
    .join("");
  updateEditSemOpts();
}
function updateEditSemOpts() {
  const pi = parseInt(document.getElementById("eProg").value);
  const yi = parseInt(document.getElementById("eYear").value);
  const prog = dbPrograms.find((p) => p.id === pi);
  const year = prog.years.find((y) => y.id === yi);
  document.getElementById("eSem").innerHTML = year.sems
    .map((s) => `<option value="${s.id}">${s.name}</option>`)
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

  // Find the original code of this course before the edit
  let originalCode = null;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === id) originalCode = c.code;
        }),
      ),
    ),
  );

  // Find other courses sharing the same original code
  const duplicates = [];
  if (originalCode) {
    dbPrograms.forEach((p) =>
      p.years.forEach((y) =>
        y.sems.forEach((s) =>
          s.courses.forEach((c) => {
            if (c.id !== id && c.code === originalCode)
              duplicates.push({
                id: c.id,
                name: c.name,
                prog: p.name,
                year: y.name,
                sem: s.name,
              });
          }),
        ),
      ),
    );
  }

  _pendingCourseEdit = {
    id,
    name,
    code,
    semId,
    duplicateIds: duplicates.map((d) => d.id),
  };

  if (duplicates.length > 0) {
    const dupList = duplicates
      .map(
        (d) =>
          `<li style="margin-bottom:6px;"><strong>${d.name}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${d.prog} › ${d.year} › ${d.sem}</span></li>`,
      )
      .join("");
    openModal(`<h2>🔁 Shared Course Detected</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;">
      <strong>${duplicates.length}</strong> other course(s) share the code <strong>${originalCode}</strong>:
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
  const { id, name, code, semId, duplicateIds } = _pendingCourseEdit;
  try {
    // Update this course with new name, code AND location
    await sb(`courses?id=eq.${id}`, "PATCH", {
      name,
      code,
      semester_id: semId,
    });
    // If syncing all, update only name + code on duplicates (keep their location)
    if (updateAll && duplicateIds.length) {
      await Promise.all(
        duplicateIds.map((did) =>
          sb(`courses?id=eq.${did}`, "PATCH", {
            name,
            code,
          }),
        ),
      );
    }
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminCourses();
    showToast(
      updateAll
        ? `Updated all ${duplicateIds.length + 1} occurrences!`
        : "Course updated!",
    );
  } catch (e) {
    showToast(e.message, true);
  } finally {
    _pendingCourseEdit = null;
  }
}

// Add Link
function openAddLinkModal(courseId) {
  openModal(`<h2>+ Add Link</h2>
  <label>Type</label><select id="lType"><option value="telegram">Telegram</option><option value="drive">Google Drive</option><option value="classroom">Google Classroom</option><option value="other">Other / External</option></select>
  <label>URL</label><input type="text" id="lUrl" placeholder="https://…"/>
  <label>Label</label><input type="text" id="lLabel" placeholder="Link 1"/>
  <label>Note (optional)</label><input type="text" id="lNote" placeholder="e.g. mail 3adi"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addLink(${courseId})">Add</button></div>`);
}
async function addLink(courseId) {
  const url = document.getElementById("lUrl").value.trim();
  if (!url) {
    showToast("URL required.", true);
    return;
  }
  const type = document.getElementById("lType").value;
  const label = document.getElementById("lLabel").value.trim() || "Link";
  const note = document.getElementById("lNote").value.trim();

  const { code, siblings } = _findSharedCourses(courseId);
  _pendingLinkOp = {
    courseId,
    url,
    type,
    label,
    note,
    siblingCourseIds: siblings.map((s) => s.id),
  };

  if (siblings.length) {
    const list = siblings
      .map(
        (s) =>
          `<li style="margin-bottom:4px;"><strong>${s.name}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${s.prog} › ${s.year} › ${s.sem}</span></li>`,
      )
      .join("");
    openModal(`<h2>🔁 Shared Course</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${siblings.length}</strong> other course(s) share the code <strong>${code}</strong>:</p>
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
  dbPrograms.forEach((p) =>
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
  const { courseId, url, type, label, note, siblingCourseIds } = _pendingLinkOp;
  try {
    await sb("links", "POST", {
      course_id: courseId,
      type,
      url,
      label,
      note,
      display_order: _getNextDisplayOrder(courseId),
    });
    if (addToAll && siblingCourseIds.length)
      await Promise.all(
        siblingCourseIds.map((sid) =>
          sb("links", "POST", {
            course_id: sid,
            type,
            url,
            label,
            note,
            display_order: _getNextDisplayOrder(sid),
          }),
        ),
      );
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminCourses();
    showToast(
      addToAll
        ? `Link added to all ${siblingCourseIds.length + 1} courses!`
        : "Link added!",
    );
  } catch (e) {
    showToast(e.message, true);
  } finally {
    _pendingLinkOp = null;
  }
}

// Edit Link
function openEditLinkModal(linkId, courseId) {
  let l;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((lk) => {
            if (lk.id === linkId) l = lk;
          }),
        ),
      ),
    ),
  );
  if (!l) return;
  openModal(`<h2>✏️ Edit Link</h2>
  <label>Type</label>
  <select id="elType">
    <option value="telegram" ${l.type === "telegram" ? "selected" : ""}>Telegram</option>
    <option value="drive" ${l.type === "drive" ? "selected" : ""}>Google Drive</option>
    <option value="classroom" ${l.type === "classroom" ? "selected" : ""}>Google Classroom</option>
    <option value="other" ${l.type === "other" ? "selected" : ""}>Other / External</option>
  </select>
  <label>URL</label><input type="text" id="elUrl" value="${l.url}"/>
  <label>Label</label><input type="text" id="elLabel" value="${l.label}"/>
  <label>Note (optional)</label><input type="text" id="elNote" value="${l.note || ""}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveLink(${linkId},${courseId})">Save</button></div>`);
}
async function saveLink(linkId, courseId) {
  const url = document.getElementById("elUrl").value.trim();
  if (!url) {
    showToast("URL required.", true);
    return;
  }
  const type = document.getElementById("elType").value;
  const label = document.getElementById("elLabel").value.trim() || "Link";
  const note = document.getElementById("elNote").value.trim();

  // Find original URL of this link before the edit
  let originalUrl = null;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) =>
          c.links.forEach((lk) => {
            if (lk.id === linkId) originalUrl = lk.url;
          }),
        ),
      ),
    ),
  );

  // Find matching links in sibling courses by original URL
  const { code, siblings } = _findSharedCourses(courseId);
  const matchingLinks = [];
  siblings.forEach((sib) =>
    sib.links.forEach((lk) => {
      if (lk.url === originalUrl)
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
    url,
    type,
    label,
    note,
    matchingLinkIds: matchingLinks.map((m) => m.id),
  };

  if (matchingLinks.length) {
    const list = matchingLinks
      .map(
        (m) =>
          `<li style="margin-bottom:4px;"><strong>${m.sibName}</strong> <span style="color:var(--muted);font-size:.8rem;">— ${m.prog} › ${m.year} › ${m.sem}</span></li>`,
      )
      .join("");
    openModal(`<h2>🔁 Shared Link</h2>
  <p style="font-size:.85rem;color:var(--muted);margin:10px 0 12px;"><strong>${matchingLinks.length}</strong> sibling course(s) have the same link (matched by URL):</p>
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
  const { linkId, url, type, label, note, matchingLinkIds } = _pendingLinkOp;
  try {
    await sb(`links?id=eq.${linkId}`, "PATCH", {
      type,
      url,
      label,
      note,
    });
    if (updateAll && matchingLinkIds.length)
      await Promise.all(
        matchingLinkIds.map((mid) =>
          sb(`links?id=eq.${mid}`, "PATCH", {
            type,
            url,
            label,
            note,
          }),
        ),
      );
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminCourses();
    showToast(
      updateAll
        ? `Updated ${matchingLinkIds.length + 1} links!`
        : "Link updated!",
    );
  } catch (e) {
    showToast(e.message, true);
  } finally {
    _pendingLinkOp = null;
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
  if (!title) {
    showToast("Title required.", true);
    return;
  }
  try {
    await sb("extra_sections", "POST", {
      title,
      icon,
      display_order: 0,
    });
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminExtra();
    showToast("Section added!");
  } catch (e) {
    showToast(e.message, true);
  }
}
function openEditExtraSectionModal(id) {
  const r = dbExtra.find((s) => s.id === id);
  if (!r) return;
  openModal(`<h2>✏️ Edit Section</h2>
  <label>Icon</label><input type="text" id="exIcon" value="${r.icon}"/>
  <label>Title</label><input type="text" id="exTitle" value="${r.title}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveExtraSection(${id})">Save</button></div>`);
}
async function saveExtraSection(id) {
  try {
    await sb(`extra_sections?id=eq.${id}`, "PATCH", {
      icon: document.getElementById("exIcon").value.trim() || "📁",
      title: document.getElementById("exTitle").value.trim(),
    });
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminExtra();
    showToast("Section updated!");
  } catch (e) {
    showToast(e.message, true);
  }
}

// Extra Links
function openAddExtraLinkModal(sectionId) {
  openModal(`<h2>+ Add Link</h2>
  <label>Type</label><select id="elxType"><option value="telegram">Telegram</option><option value="drive">Google Drive</option><option value="classroom">Google Classroom</option><option value="other">Other / External</option></select>
  <label>URL</label><input type="text" id="elxUrl" placeholder="https://…"/>
  <label>Label</label><input type="text" id="elxLabel" placeholder="Link 1"/>
  <label>Note (optional)</label><input type="text" id="elxNote"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="addExtraLink(${sectionId})">Add</button></div>`);
}
async function addExtraLink(sectionId) {
  const url = document.getElementById("elxUrl").value.trim();
  if (!url) {
    showToast("URL required.", true);
    return;
  }
  try {
    await sb("extra_links", "POST", {
      section_id: sectionId,
      type: document.getElementById("elxType").value,
      url,
      label: document.getElementById("elxLabel").value.trim() || "Link",
      note: document.getElementById("elxNote").value.trim(),
      display_order: 0,
    });
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminExtra();
    showToast("Link added!");
  } catch (e) {
    showToast(e.message, true);
  }
}
function openEditExtraLinkModal(linkId, sectionId) {
  const sec = dbExtra.find((s) => s.id === sectionId);
  const l = sec && sec.links.find((lk) => lk.id === linkId);
  if (!l) return;
  openModal(`<h2>✏️ Edit Link</h2>
  <label>Type</label>
  <select id="elxType">
    <option value="telegram" ${l.type === "telegram" ? "selected" : ""}>Telegram</option>
    <option value="drive" ${l.type === "drive" ? "selected" : ""}>Google Drive</option>
    <option value="classroom" ${l.type === "classroom" ? "selected" : ""}>Google Classroom</option>
    <option value="other" ${l.type === "other" ? "selected" : ""}>Other / External</option>
  </select>
  <label>URL</label><input type="text" id="elxUrl" value="${l.url}"/>
  <label>Label</label><input type="text" id="elxLabel" value="${l.label}"/>
  <label>Note (optional)</label><input type="text" id="elxNote" value="${l.note || ""}"/>
  <div class="modal-actions"><button class="btn btn-ghost" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="saveExtraLink(${linkId})">Save</button></div>`);
}
async function saveExtraLink(id) {
  const url = document.getElementById("elxUrl").value.trim();
  if (!url) {
    showToast("URL required.", true);
    return;
  }
  try {
    await sb(`extra_links?id=eq.${id}`, "PATCH", {
      type: document.getElementById("elxType").value,
      url,
      label: document.getElementById("elxLabel").value.trim() || "Link",
      note: document.getElementById("elxNote").value.trim(),
    });
    closeModal();
    await trackVisit();
    loadAll();
    renderAdminExtra();
    showToast("Link updated!");
  } catch (e) {
    showToast(e.message, true);
  }
}