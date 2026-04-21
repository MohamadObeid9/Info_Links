
// ===================== VIEWS =====================

// ── Clean URLs (History API) ────────────────────────────────────────────────
const VALID_VIEWS = ["home", "report-submit", "feedback", "admin-gate", "admin"];

function _getPathView() {
  const path = window.location.pathname.replace("/", "");
  return VALID_VIEWS.includes(path) ? path : "home";
}

function showView(v) {
  // 1. Security Guard: Prevent direct access to #admin if not logged in
  if (v === "admin" && !AppState.adminLoggedIn) {
    console.warn("Unauthorized access attempt to admin dashboard.");
    v = "admin-gate";
  }

  // Handle the logic for admin-gate redirect
  if (v === "admin-gate" && AppState.adminLoggedIn) v = "admin";

  if (v === "feedback") updateStarDisplay();

  // 2. UI Updates
  document.querySelectorAll(".view").forEach((el) => el.classList.remove("active"));
  document.querySelectorAll(".nav-btn").forEach((b) => {
    b.classList.remove("active");
    if (b.dataset.view === v) b.classList.add("active");
  });

  const viewEl = document.getElementById("view-" + v);
  if (viewEl) viewEl.classList.add("active");

  // 3. Update URL (Clean)
  const newPath = "/" + (v === "home" ? "" : v);
  if (window.location.pathname !== newPath) {
    window.history.pushState({ view: v }, "", newPath);
  }

  // 4. Trigger logic
  if (v === "admin") {
    renderAdminContent();
    loadReportsBadges();
  }
}

// Restore view on back-button
window.addEventListener("popstate", (event) => {
  const v = event.state?.view || _getPathView();
  showView(v);
});

// ===================== REPORT & CONTRIB =====================

// When a course is selected in the report form, populate its links dropdown
function onReportCourseChange() {
  const courseInput = document.getElementById("rCourse").value.trim().toLowerCase();
  const linkSel = document.getElementById("rLink");

  if (!courseInput) {
    if (!linkSel.disabled) {
      linkSel.innerHTML = '<option value="">Select a link…</option>';
      linkSel.disabled = true;
    }
    return;
  }

  // Find the course in the data tree by name or code
  let course = null;
  AppState.dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.name.toLowerCase() === courseInput || c.code.toLowerCase() === courseInput || `${c.name} (${c.code})`.toLowerCase() === courseInput) {
            course = c;
          }
        }),
      ),
    ),
  );

  if (course && course.links.length) {
    linkSel.innerHTML = '<option value="">Select a link…</option>';
    course.links.forEach((l) => {
      const opt = document.createElement("option");
      opt.value = l.label;
      opt.textContent = l.label;
      linkSel.appendChild(opt);
    });
    linkSel.disabled = false;
  } else {
    if (!linkSel.disabled) {
      linkSel.innerHTML = '<option value="">Select a link…</option>';
      linkSel.disabled = true;
    }
  }
}

async function submitReport() {
  const btn = document.getElementById("submitReportBtn");
  const courseName = document.getElementById("rCourse").value.trim();
  const link = document.getElementById("rLink").value;
  const desc = document.getElementById("rDesc").value.trim();
  if (!courseName || !desc) {
    showToast("Please select a course and describe the issue.", true);
    return;
  }
  setBtnLoading(btn, true, "Submitting…");
  try {
    const res = await fetch("/api/reports", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        course_name: courseName,
        link_url: link || "",
        description: desc,
      })
    });
    if (!res.ok) throw new Error("Server error");
    document.getElementById("rCourse").value = "";
    onReportCourseChange(); // reset link dropdown
    document.getElementById("rDesc").value = "";
    showToast("Report submitted! Thank you.");
  } catch (e) {
    showToast("Failed to submit: " + e.message, true);
  } finally {
    setBtnLoading(btn, false);
  }
}

async function submitContribution() {
  const btn = document.getElementById("submitContribBtn");
  const course = document.getElementById("cCourse").value.trim();
  const link = document.getElementById("cLink").value.trim();
  const linkType = document.getElementById("cType").value;
  const note = document.getElementById("cNote").value.trim();
  if (!course || !link) {
    showToast("Please fill in course and link.", true);
    return;
  }
  setBtnLoading(btn, true, "Submitting…");
  try {
    const finalNote = linkType ? `[Type:${linkType}] ${note}` : note;
    const res = await fetch("/api/contributions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        course_name: course,
        link_url: link,
        link_type: linkType,
        note: finalNote.trim(),
      })
    });
    if (!res.ok) throw new Error("Server error");
    document.getElementById("cCourse").value = "";
    document.getElementById("cLink").value = "";
    document.getElementById("cType").value = "";
    document.getElementById("cNote").value = "";
    showToast("Contribution submitted! Thank you.");
  } catch (e) {
    showToast("Failed to submit: " + e.message, true);
  } finally {
    setBtnLoading(btn, false);
  }
}

window.showView = showView; 
window.onReportCourseChange = onReportCourseChange; 
window.submitReport = submitReport; 
window.submitContribution = submitContribution;
