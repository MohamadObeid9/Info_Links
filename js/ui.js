// ===================== THEME =====================
function toggleTheme() {
  isDark = !isDark;
  document.documentElement.setAttribute(
    "data-theme",
    isDark ? "dark" : "light",
  );
  document.getElementById("themeBtn").textContent = isDark ? "🌙" : "☀️";
}

// ===================== MOBILE =====================
function toggleMobileMenu() {
  document.getElementById("hamburgerBtn").classList.toggle("open");
  document.getElementById("navLinks").classList.toggle("mobile-open");
}
function toggleFilters() {
  document.getElementById("filterToggleBtn").classList.toggle("open");
  document.getElementById("filtersCollapsible").classList.toggle("open");
}
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("nav-btn")) {
    document.getElementById("hamburgerBtn").classList.remove("open");
    document.getElementById("navLinks").classList.remove("mobile-open");
  }
});

// Helper: find all courses that share the same code as the given courseId
function _findSharedCourses(courseId) {
  let targetCode = null;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === courseId) targetCode = c.code;
        }),
      ),
    ),
  );
  if (!targetCode) return { code: null, siblings: [] };
  const siblings = [];
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id !== courseId && c.code === targetCode)
            siblings.push({
              ...c,
              prog: p.name,
              year: y.name,
              sem: s.name,
            });
        }),
      ),
    ),
  );
  return { code: targetCode, siblings };
}

// ===================== BADGE =====================
function getLinkBadge(type) {
  if (type === "telegram") return '<span class="link-badge badge-tg">TG</span>';
  if (type === "drive") return '<span class="link-badge badge-drive">GD</span>';
  if (type === "classroom")
    return '<span class="link-badge badge-classroom">GC</span>';
  if (type === "other") return '<span class="link-badge badge-other">OT</span>';
  return '<span class="link-badge badge-other">OT</span>';
}