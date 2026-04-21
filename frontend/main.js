// --- EVENT ROUTER (Professional Event Delegation) ---
document.addEventListener('click', (e) => {
  const target = e.target;

  // 1. View Switching
  const view = target.closest('[data-view]')?.dataset.view;
  if (view) {
    window.showView(view);
    return;
  }

  // 2. Action Routing
  const action = target.closest('[data-action]')?.dataset.action;
  if (action) {
    const actionEl = target.closest('[data-action]');
    switch (action) {
      case 'toggleMobileMenu': window.toggleMobileMenu(); break;
      case 'toggleTheme': window.toggleTheme(); break;
      case 'toggleImportantNote': window.toggleImportantNote(); break;
      case 'toggleFilters': window.toggleFilters(); break;
      case 'submitReport': window.submitReport(); break;
      case 'submitContribution': window.submitContribution(); break;
      case 'submitFeedback': window.submitFeedback(); break;
      case 'logout': window.logout(); break;
      case 'setNoteLang': 
        e.stopPropagation(); 
        window.setNoteLang(actionEl?.dataset.lang);
        break;
    }
    return;
  }

  // 3. Admin Tabs
  const adminTabName = target.closest('[data-admin-tab]')?.dataset.adminTab;
  if (adminTabName) {
    window.adminTab(adminTabName);
    return;
  }

  // 4. Rating Stars
  const rating = target.closest('[data-rating]')?.dataset.rating;
  if (rating) {
    window.setRating(parseInt(rating, 10));
    return;
  }
});

// --- SEARCH ROUTER ---
document.addEventListener('input', (e) => {
  if (e.target.id === 'searchInput') {
    window.onSearch();
  }
  if (e.target.id === 'rCourse') {
    window.onReportCourseChange();
  }
});

// --- HOVER ROUTER (For Feedback Stars) ---
document.addEventListener('mouseover', (e) => {
  const rating = e.target.closest('[data-rating]')?.dataset.rating;
  if (rating) window.handleStarHover(parseInt(rating, 10));
});
document.addEventListener('mouseout', (e) => {
  if (e.target.closest('[data-rating]')) window.clearStarHover();
});

// --- INITIALIZATION ---
async function init() {
  try {
    await window.initApp();
  } catch (err) {
    console.error("Critical: App initialization failed:", err);
  }
}

window.addEventListener("DOMContentLoaded", init);
