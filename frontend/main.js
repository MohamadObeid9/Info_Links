import "./js/state.js";
import "./js/cache.js";
import "./js/supabase.js";
import "./js/skeleton.js";
import "./js/ui.js";
import "./js/home.js";
import "./js/data.js";
import "./js/feedback.js";
import "./js/export.js";
import "./js/admin.js";
import "./js/views.js";
import "./js/modals.js";

// --- EVENT ROUTER (Professional Event Delegation) ---
document.addEventListener("click", (e) => {
  const target = e.target;

  // Copy-link button (must run before the link-item handler below)
  const copyBtn = target.closest(".copy-btn");
  if (copyBtn) {
    e.preventDefault();
    e.stopPropagation();
    const copyUrl = copyBtn.closest("[data-url]")?.dataset.url;
    if (copyUrl) window.copyLink(copyUrl);
    return;
  }

  // External link item → confirmation modal (URL read from dataset, never inline JS)
  const linkItem = target.closest(".link-item");
  if (linkItem) {
    e.preventDefault();
    const idAttr = linkItem.dataset.linkId;
    const linkId = idAttr ? parseInt(idAttr, 10) : null;
    const linkKind = linkItem.dataset.linkKind || "link";
    window.confirmLink(linkId, linkItem.dataset.url || null, linkKind);
    return;
  }

  const view = target.closest("[data-view]")?.dataset.view;
  if (view) {
    window.showView(view);
    return;
  }

  const action = target.closest("[data-action]")?.dataset.action;
  if (action) {
    const actionEl = target.closest("[data-action]");
    switch (action) {
      case "toggleMobileMenu":
        window.toggleMobileMenu();
        break;
      case "toggleTheme":
        window.toggleTheme();
        break;
      case "toggleFilters":
        window.toggleFilters();
        break;
      case "submitReport":
        window.submitReport();
        break;
      case "submitContribution":
        window.submitContribution();
        break;
      case "submitFeedback":
        window.submitFeedback();
        break;
      case "logout":
        window.logout();
        break;
      case "checkLogin":
        window.checkLogin();
        break;
      case "openAddCourseModal":
        window.openAddCourseModal();
        break;
      case "exportData":
        window.exportData();
        break;
    }
    return;
  }

  const adminTabName = target.closest("[data-admin-tab]")?.dataset.adminTab;
  if (adminTabName) {
    window.adminTab(adminTabName);
    return;
  }

  const rating = target.closest("[data-rating]")?.dataset.rating;
  if (rating) {
    window.setRating(parseInt(rating, 10));
    return;
  }
});

document.addEventListener("input", (e) => {
  if (e.target.id === "searchInput") {
    window.onSearch();
  }
  if (e.target.id === "rCourse") {
    window.onReportCourseChange();
  }
});

document.addEventListener("keydown", (e) => {
  if (e.key === "Enter" && e.target.id === "adminPass") {
    e.preventDefault();
    window.checkLogin();
  }
});

document.addEventListener("mouseover", (e) => {
  const rating = e.target.closest("[data-rating]")?.dataset.rating;
  if (rating) window.handleStarHover(parseInt(rating, 10));
});
document.addEventListener("mouseout", (e) => {
  if (e.target.closest("[data-rating]")) window.clearStarHover();
});

async function init() {
  try {
    await window.initApp();
  } catch (err) {
    console.error("Critical: App initialization failed:", err);
  }
}

window.addEventListener("DOMContentLoaded", init);
