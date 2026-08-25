// ===================== STATE =====================
// All mutable application state lives here as a single object.
// Access via AppState.xxx; mutate directly: AppState.xxx = yyy.

function _isTokenValid(token) {
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 > Date.now();
  } catch {
    return false;
  }
}

const _rawToken = localStorage.getItem("infolinks_token");
const _validToken = _isTokenValid(_rawToken) ? _rawToken : null;
if (_rawToken && !_validToken) localStorage.removeItem("infolinks_token");

// Student session token — completely separate from the admin token above.
const STUDENT_TOKEN_KEY = "infolinks_student_token";
// Last known registered student id, so the favorites cache can paint before
// GET /api/users/me resolves.
const STUDENT_UID_KEY = "infolinks_student_uid";
const _rawStudentToken = localStorage.getItem(STUDENT_TOKEN_KEY);
const _validStudentToken = _isTokenValid(_rawStudentToken) ? _rawStudentToken : null;
if (_rawStudentToken && !_validStudentToken) {
  localStorage.removeItem(STUDENT_TOKEN_KEY);
  localStorage.removeItem(STUDENT_UID_KEY);
}
const _storedStudentId = _validStudentToken
  ? parseInt(localStorage.getItem(STUDENT_UID_KEY) || "", 10) || null
  : null;

const AppState = {
  sbToken: _validToken,
  adminLoggedIn: !!_validToken,
  studentToken: _validStudentToken,
  studentUser: null,
  studentUserId: _storedStudentId,
  adminStudentId: null,
  currentAdminTab: "courses",
  adminSearch: "",
  adminFilterProg: "all",
  adminFilterYear: "all",
  adminFilterSem: "all",
  _pendingCourseEdit: null,
  _pendingLinkOp: null,
  isDark: false,
  currentProg: null,
  currentYear: "all",
  currentSem: "all",
  // Phone browse: "program" picker, "year" (year+semester), or "list".
  mobileStep: "program",
  dbPrograms: [],
  analyticsRange: "30",
  analyticsChartSeries: "visitors",
  analyticsVisitorsSort: "clicks",
  analyticsVisitorsOffset: 0,
  analyticsLinksTab: "today",
  analyticsTopLinksExpanded: false,
  analyticsTopLinksTodayExpanded: false,
  analyticsTopUsersExpanded: false,
  analyticsTopCoursesExpanded: false,
  analyticsTopFavoritesExpanded: false,
  analyticsZeroCoursesExpanded: false,
  analyticsZeroLinksExpanded: false,
  analyticsSearchExpanded: false,
  dbExtra: [],
  courseById: new Map(),
  linkById: new Map(),
  favorites: new Set(), // Set of course IDs (strings from localStorage)
};

// ── Favorites ────────────────────────────────────────────────────────────────
// Favorites live server-side; localStorage is only a per-user cache so the star
// state paints instantly on reload. The legacy shared key is dropped so one
// student never inherits another's favorites on the same browser.
localStorage.removeItem("infolinks_favorites");
localStorage.removeItem("infolinks_mobile_browse");

function _favoritesCacheKey() {
  const id = AppState.studentUser?.id ?? AppState.studentUserId;
  return id ? `infolinks_favorites_u${id}` : null;
}

function saveFavorites() {
  const key = _favoritesCacheKey();
  if (!key) return;
  try {
    localStorage.setItem(key, JSON.stringify([...AppState.favorites]));
  } catch (e) { }
}

/** Paint favorites from the per-user cache (called before /users/me resolves). */
function loadFavoritesCache() {
  const key = _favoritesCacheKey();
  if (!key) return false;
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return false;
    AppState.favorites = new Set(JSON.parse(raw).map(String));
    return true;
  } catch (e) {
    return false;
  }
}

/** Replace the local set with the authoritative server list. */
function setFavorites(courseIds) {
  AppState.favorites = new Set((courseIds || []).map(String));
  saveFavorites();
}

function clearFavorites() {
  const key = _favoritesCacheKey();
  AppState.favorites = new Set();
  if (key) {
    try {
      localStorage.removeItem(key);
    } catch (e) { }
  }
}

function toggleFavorite(courseId) {
  const key = String(courseId);
  if (AppState.favorites.has(key)) {
    AppState.favorites.delete(key);
  } else {
    AppState.favorites.add(key);
  }
  saveFavorites();
}

// ── Backward-compat shims so legacy code reading bare vars still works ────────
// (These will be cleaned up gradually; new code should use AppState.xxx)
Object.defineProperties(window, {
  sbToken: { get: () => AppState.sbToken, set: v => { AppState.sbToken = v; } },
  adminLoggedIn: { get: () => AppState.adminLoggedIn, set: v => { AppState.adminLoggedIn = v; } },
  currentAdminTab: { get: () => AppState.currentAdminTab, set: v => { AppState.currentAdminTab = v; } },
  adminSearch: { get: () => AppState.adminSearch, set: v => { AppState.adminSearch = v; } },
  adminFilterProg: { get: () => AppState.adminFilterProg, set: v => { AppState.adminFilterProg = v; } },
  adminFilterYear: { get: () => AppState.adminFilterYear, set: v => { AppState.adminFilterYear = v; } },
  adminFilterSem: { get: () => AppState.adminFilterSem, set: v => { AppState.adminFilterSem = v; } },
  _pendingCourseEdit: { get: () => AppState._pendingCourseEdit, set: v => { AppState._pendingCourseEdit = v; } },
  _pendingLinkOp: { get: () => AppState._pendingLinkOp, set: v => { AppState._pendingLinkOp = v; } },
  isDark: { get: () => AppState.isDark, set: v => { AppState.isDark = v; } },
  currentProg: { get: () => AppState.currentProg, set: v => { AppState.currentProg = v; } },
  currentYear: { get: () => AppState.currentYear, set: v => { AppState.currentYear = v; } },
  currentSem: { get: () => AppState.currentSem, set: v => { AppState.currentSem = v; } },
  dbPrograms: { get: () => AppState.dbPrograms, set: v => { AppState.dbPrograms = v; } },
  analyticsRange: { get: () => AppState.analyticsRange, set: v => { AppState.analyticsRange = v; } },
  dbExtra: { get: () => AppState.dbExtra, set: v => { AppState.dbExtra = v; } },
});

window.AppState = AppState;
export {
  AppState,
  STUDENT_TOKEN_KEY,
  STUDENT_UID_KEY,
  _isTokenValid as isJwtValid,
  saveFavorites,
  loadFavoritesCache,
  setFavorites,
  clearFavorites,
  toggleFavorite,
};