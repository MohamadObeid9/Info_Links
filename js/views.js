// ===================== VIEWS =====================
function showView(v) {
  document
    .querySelectorAll(".view")
    .forEach((el) => el.classList.remove("active"));
  document
    .querySelectorAll(".nav-btn")
    .forEach((b) => b.classList.remove("active"));
  if (v === "admin-gate" && adminLoggedIn) v = "admin";
  if (v === "feedback") {
    updateStarDisplay();
  }
  document.getElementById("view-" + v).classList.add("active");
  if (v === "admin") {
    renderAdminContent();
    loadReportsBadges();
  }
}

// ===================== REPORT & CONTRIB =====================
async function submitReport() {
  const course = document.getElementById("rCourse").value.trim();
  const link = document.getElementById("rLink").value.trim();
  const desc = document.getElementById("rDesc").value.trim();
  if (!course || !desc) {
    showToast("Please fill in course and description.", true);
    return;
  }
  try {
    await sb("reports", "POST", {
      course_name: course,
      link_url: link,
      description: desc,
      status: "open",
      reply: "",
    });
    document.getElementById("rCourse").value = "";
    document.getElementById("rLink").value = "";
    document.getElementById("rDesc").value = "";
    showToast("Report submitted! Thank you.");
  } catch (e) {
    showToast("Failed to submit: " + e.message, true);
  }
}

async function submitContribution() {
  const course = document.getElementById("cCourse").value.trim();
  const link = document.getElementById("cLink").value.trim();
  const note = document.getElementById("cNote").value.trim();
  if (!course || !link) {
    showToast("Please fill in course and link.", true);
    return;
  }
  try {
    await sb("contributions", "POST", {
      course_name: course,
      link_url: link,
      note,
      status: "pending",
    });
    document.getElementById("cCourse").value = "";
    document.getElementById("cLink").value = "";
    document.getElementById("cNote").value = "";
    showToast("Contribution submitted! Thank you.");
  } catch (e) {
    showToast("Failed to submit: " + e.message, true);
  }
}