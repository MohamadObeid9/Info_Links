// ===================== STATE =====================
let sbToken = null;
let adminLoggedIn = false;
let currentAdminTab = "courses";
let adminSearch = "";
let adminFilterProg = "all";
let adminFilterYear = "all";
let adminFilterSem = "all";
let _pendingCourseEdit = null;
let _pendingLinkOp = null;
let isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
let currentProg = null;
let currentYear = "all";
let currentSem = "all";
let dbPrograms = [];
let analyticsRange = "30";
let dbExtra = [];

document.documentElement.setAttribute("data-theme", isDark ? "dark" : "light");
document.getElementById("themeBtn").textContent = isDark ? "🌙" : "☀️";