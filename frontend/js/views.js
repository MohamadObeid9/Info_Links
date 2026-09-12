import { AppState } from "./state.js";
import { apiRequest } from "./supabase.js";
import { setBtnLoading } from "./ui.js";
import { showToast } from "./export.js";
import { renderAdminContent } from "./admin.js";
import { loadReportsBadges } from "./data.js";
import { updateStarDisplay } from "./feedback.js";
import { initHierarchyPicker, selectedCourse } from "./hierarchy-picker.js";
import {
  encodeContributionNote,
  _readContentTypeCheckboxes,
} from "./modals.js";

// ===================== VIEWS =====================

// ── Clean URLs (History API) ────────────────────────────────────────────────
const VALID_VIEWS = ["home", "report-submit", "feedback", "about", "admin-gate", "admin"];

function _getPathView() {
  const path = window.location.pathname.replace("/", "");
  return VALID_VIEWS.includes(path) ? path : "home";
}

function _expireTokenIfNeeded() {
  if (!AppState.sbToken) return;
  try {
    const p = JSON.parse(atob(AppState.sbToken.split('.')[1]));
    if (p.exp * 1000 <= Date.now()) {
      localStorage.removeItem("infolinks_token");
      AppState.sbToken = null;
      AppState.adminLoggedIn = false;
    }
  } catch {
    localStorage.removeItem("infolinks_token");
    AppState.sbToken = null;
    AppState.adminLoggedIn = false;
  }
}

function showView(v) {
  // Always check token expiry before any view change
  _expireTokenIfNeeded();

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
  if (v === "report-submit") {
    refreshReportContributePickers();
  }
}

// Restore view on back-button
window.addEventListener("popstate", (event) => {
  const v = event.state?.view || _getPathView();
  showView(v);
});

// ===================== REPORT & CONTRIB =====================

function refreshReportContributePickers() {
  if (document.getElementById("rProg")) initHierarchyPicker("r", { includeCourse: true });
  if (document.getElementById("cProg")) initHierarchyPicker("c", { includeCourse: true });
}

function onHierarchyCourseChange(prefix) {
  if (prefix === "c") {
    syncContributeLinkDetails();
    return;
  }
  if (prefix !== "r") return;
  syncReportLinkDetails();
}

function syncReportLinkDetails() {
  const details = document.getElementById("rReportDetails");
  const placeholder = document.getElementById("rReportPlaceholder");
  const linkSel = document.getElementById("rLink");
  const linkStep = document.getElementById("rLinkStep");
  if (!details || !linkSel) return;

  const course = selectedCourse("r");
  const show = !!course;
  details.hidden = !show;
  if (placeholder) placeholder.hidden = show;

  if (!show) {
    linkSel.innerHTML = `<option value="">Select a link…</option>`;
    linkSel.disabled = true;
    if (linkStep) linkStep.hidden = true;
    const desc = document.getElementById("rDesc");
    if (desc) desc.value = "";
    return;
  }

  if (course.links?.length) {
    linkSel.innerHTML = `<option value="">Select a link…</option>`;
    course.links.forEach((l) => {
      const opt = document.createElement("option");
      opt.value = l.label || l.url || "";
      opt.textContent = l.label || l.url || "Link";
      linkSel.appendChild(opt);
    });
    linkSel.disabled = false;
    if (linkStep) linkStep.hidden = false;
  } else {
    linkSel.innerHTML = `<option value="">Select a link…</option>`;
    linkSel.disabled = true;
    if (linkStep) linkStep.hidden = true;
  }

  queueMicrotask(() => {
    document.getElementById("rDesc")?.focus();
  });
}

function syncContributeLinkDetails() {
  const details = document.getElementById("cLinkDetails");
  const placeholder = document.getElementById("cContribPlaceholder");
  if (!details) return;
  const course = selectedCourse("c");
  const show = !!course;
  details.hidden = !show;
  if (placeholder) placeholder.hidden = show;
  if (!show) {
    const link = document.getElementById("cLink");
    const type = document.getElementById("cType");
    const note = document.getElementById("cNote");
    if (link) link.value = "";
    if (type) type.value = "";
    if (note) note.value = "";
    document.querySelectorAll('input[name="cct"]').forEach((el) => {
      el.checked = false;
    });
    return;
  }
  queueMicrotask(() => {
    document.getElementById("cLink")?.focus();
  });
}

async function submitReport() {
  if (!window.requireStudent(submitReport)) return;
  const btn = document.getElementById("submitReportBtn");
  const course = selectedCourse("r");
  const link = document.getElementById("rLink").value;
  const desc = document.getElementById("rDesc").value.trim();
  if (!course || !desc) {
    showToast("Please select a course and describe the issue.", true);
    return;
  }
  const courseName = course.code ? `${course.name} (${course.code})` : course.name;
  setBtnLoading(btn, true, "Submitting…");
  try {
    await apiRequest("/api/reports", {
      method: "POST",
      body: {
        course_name: courseName,
        link_url: link || "",
        description: desc,
      },
    });
    initHierarchyPicker("r", { includeCourse: true });
    syncReportLinkDetails();
    showToast("Report submitted! Thank you.");
  } catch (e) {
    if (window.handleStudentAuthError?.(e, submitReport)) return;
    showToast("Failed to submit: " + e.message, true);
  } finally {
    setBtnLoading(btn, false);
  }
}

async function submitContribution() {
  if (!window.requireStudent(submitContribution)) return;
  const btn = document.getElementById("submitContribBtn");
  const course = selectedCourse("c");
  const link = document.getElementById("cLink").value.trim();
  const linkType = document.getElementById("cType").value;
  const note = document.getElementById("cNote").value.trim();
  const contentTypes = _readContentTypeCheckboxes("cct");
  if (!course || !link) {
    showToast("Please select a course and enter a link.", true);
    return;
  }
  try {
    const p = new URL(link);
    if (p.protocol !== "http:" && p.protocol !== "https:") throw new Error();
  } catch {
    showToast("Please enter a valid URL (https://… or http://…).", true);
    return;
  }
  const courseName = course.code ? `${course.name} (${course.code})` : course.name;
  setBtnLoading(btn, true, "Submitting…");
  try {
    const finalNote = encodeContributionNote({
      linkType,
      contentTypes,
      note,
    });
    await apiRequest("/api/contributions", {
      method: "POST",
      body: {
        course_name: courseName,
        link_url: link,
        link_type: linkType,
        note: finalNote,
      },
    });
    initHierarchyPicker("c", { includeCourse: true });
    document.getElementById("cLink").value = "";
    document.getElementById("cType").value = "";
    document.getElementById("cNote").value = "";
    document.querySelectorAll('input[name="cct"]').forEach((el) => {
      el.checked = false;
    });
    syncContributeLinkDetails();
    showToast("Contribution submitted! Thank you.");
  } catch (e) {
    if (window.handleStudentAuthError?.(e, submitContribution)) return;
    showToast("Failed to submit: " + e.message, true);
  } finally {
    setBtnLoading(btn, false);
  }
}

window.showView = showView;
window.onHierarchyCourseChange = onHierarchyCourseChange;
window.submitReport = submitReport;
window.submitContribution = submitContribution;
window.refreshReportContributePickers = refreshReportContributePickers;

export {
  showView,
  onHierarchyCourseChange,
  submitReport,
  submitContribution,
  refreshReportContributePickers,
};
