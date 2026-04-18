// ===================== VIEWS =====================

// ── Hash-based routing ─────────────────────────────────────────────────────
const VALID_VIEWS = ["home", "report-submit", "feedback", "admin-gate", "admin"];

function _getHashView() {
  const hash = window.location.hash.replace("#", "");
  return VALID_VIEWS.includes(hash) ? hash : "home";
}

function showView(v) {
  document.querySelectorAll(".view").forEach((el) => el.classList.remove("active"));
  document.querySelectorAll(".nav-btn").forEach((b) => b.classList.remove("active"));

  if (v === "admin-gate" && AppState.adminLoggedIn) v = "admin";
  if (v === "feedback") updateStarDisplay();

  document.getElementById("view-" + v).classList.add("active");

  // Update hash for deep-linking (avoid pushing duplicate history entries)
  const newHash = "#" + v;
  if (window.location.hash !== newHash) {
    window.location.hash = newHash;
  }

  if (v === "admin") {
    renderAdminContent();
    loadReportsBadges();
  }
}

// Restore view from hash on load / back-button
window.addEventListener("hashchange", () => {
  const v = _getHashView();
  // Only switch if different from current visible view
  const active = document.querySelector(".view.active");
  if (active && active.id !== "view-" + v) {
    showView(v);
  }
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
    await sb("reports", "POST", {
      course_name: courseName,
      link_url: link || "",
      description: desc,
      status: "open",
      reply: "",
    });
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
    await sb("contributions", "POST", {
      course_name: course,
      link_url: link,
      note: finalNote.trim(),
      status: "pending",
    });
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