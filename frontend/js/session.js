// ===================== STUDENT SESSION =====================
// Student identity is name + last name + number (no password). Every browser
// gets a guest session on first load; registering claims that guest row so
// pre-signup activity stays attached to the same student.
import {
  AppState,
  STUDENT_TOKEN_KEY,
  STUDENT_UID_KEY,
  loadFavoritesCache,
  setFavorites,
  clearFavorites,
} from "./state.js";
import { apiRequest, formatApiError, logApiError } from "./supabase.js";
import { openModal, closeModal } from "./modals.js";
import { esc, setBtnLoading } from "./ui.js";
import { showToast } from "./export.js";

// Action to replay once the visitor finishes signing up / signing in.
let _pendingAction = null;
// In-flight bootstrap, so gated clicks during startup wait instead of
// prompting a student who is already registered.
let _bootstrapping = null;

function studentHandle() {
  const u = AppState.studentUser;
  if (!u || u.is_guest) return null;
  return u.handle || `${u.first_name}_${u.last_name}_${u.number}`;
}

function isRegisteredStudent() {
  return !!AppState.studentUser && AppState.studentUser.is_guest === false;
}

// ── Token / identity plumbing ───────────────────────────────────────────────
function _setStudentToken(token) {
  AppState.studentToken = token;
  try {
    localStorage.setItem(STUDENT_TOKEN_KEY, token);
  } catch (e) { }
}

function _clearStudentToken() {
  AppState.studentToken = null;
  AppState.studentUser = null;
  try {
    localStorage.removeItem(STUDENT_TOKEN_KEY);
  } catch (e) { }
}

/** Drop the cached favorites + remembered id of whoever was signed in. */
function _forgetStudentIdentity() {
  clearFavorites();
  AppState.studentUserId = null;
  try {
    localStorage.removeItem(STUDENT_UID_KEY);
  } catch (e) { }
}

function applyStudentUser(user) {
  const previousId = AppState.studentUser?.id ?? AppState.studentUserId;
  if (previousId && user?.id && previousId !== user.id) _forgetStudentIdentity();

  AppState.studentUser = user || null;
  if (user && !user.is_guest) {
    AppState.studentUserId = user.id;
    try {
      localStorage.setItem(STUDENT_UID_KEY, String(user.id));
    } catch (e) { }
    setFavorites(user.favorite_course_ids);
  } else {
    _forgetStudentIdentity();
  }
  renderStudentBanner();
  repaintFavoriteStars();
  // Content may already be on screen (session and /api/content race). Re-paint
  // so link hrefs match the session: real URLs for students, "#" for guests.
  if ((AppState.dbPrograms && AppState.dbPrograms.length) || (AppState.dbExtra && AppState.dbExtra.length)) {
    window.renderCourses?.();
    window.renderExtra?.();
  }
}

async function createGuestSession() {
  const data = await apiRequest("/api/users/guest", { method: "POST" });
  if (!data?.token) throw new Error("Guest session response is missing a token");
  _setStudentToken(data.token);
}

async function refreshStudentProfile() {
  try {
    const user = await apiRequest("/api/users/me");
    applyStudentUser(user);
    return user;
  } catch (err) {
    if (err?.status === 401) {
      await resetToGuest();
      return null;
    }
    throw err;
  }
}

/** Expired / rejected student token: forget it and start a fresh guest. */
async function resetToGuest() {
  _forgetStudentIdentity();
  _clearStudentToken();
  renderStudentBanner();
  repaintFavoriteStars();
  try {
    await createGuestSession();
  } catch (err) {
    logApiError(err, "guestBootstrap");
  }
}

async function bootstrapStudentSession() {
  if (AppState.adminLoggedIn) {
    renderStudentBanner();
    return;
  }
  if (_bootstrapping) return _bootstrapping;

  _bootstrapping = (async () => {
    try {
      if (AppState.studentUserId) {
        loadFavoritesCache();
        repaintFavoriteStars();
      }
      renderStudentBanner();
      if (!AppState.studentToken) await createGuestSession();
      if (AppState.studentToken) await refreshStudentProfile();
    } catch (err) {
      logApiError(err, "studentSession");
    }
  })();

  try {
    await _bootstrapping;
  } finally {
    _bootstrapping = null;
  }
}

function onStudentTokenRejected() {
  if (_bootstrapping) return;
  resetToGuest().catch((err) => logApiError(err, "guestBootstrap"));
}

// ── Gating ──────────────────────────────────────────────────────────────────
/**
 * True when the visitor may perform a registered-only action. Otherwise the
 * signup/login modal opens and `retry` runs after a successful sign-in.
 */
function requireStudent(retry) {
  if (AppState.adminLoggedIn) return true;
  if (isRegisteredStudent()) return true;

  if (_bootstrapping) {
    _bootstrapping.then(() => {
      if (isRegisteredStudent()) {
        if (typeof retry === "function") retry();
      } else {
        promptStudentAuth({ retry });
      }
    });
    return false;
  }

  promptStudentAuth({ retry });
  return false;
}

/** 401/403 on a gated request → recover the session and prompt. */
function handleStudentAuthError(err, retry) {
  const status = err?.status;
  if (status !== 401 && status !== 403) return false;
  if (status === 401) resetToGuest().catch((e) => logApiError(e, "guestBootstrap"));
  promptStudentAuth({ retry });
  return true;
}

// ── Signup / login modal ────────────────────────────────────────────────────
function _randomNumber() {
  return Math.floor(Math.random() * 100) + 1;
}

/** Next number to suggest after a collision (55 → 65, wrapping past 100). */
function _suggestNumber(taken) {
  const next = Number(taken) + 10;
  return next > 100 ? next - 100 : next;
}

function promptStudentAuth({ retry = null, mode = "signup" } = {}) {
  _pendingAction = typeof retry === "function" ? retry : null;
  _renderAuthModal({ mode });
}

function _renderAuthModal({ mode = "signup", error = "", values = {} } = {}) {
  const isSignup = mode !== "signin";
  const first = values.first_name || "";
  const last = values.last_name || "";
  const number = values.number || (isSignup ? String(_randomNumber()) : "");

  openModal(`<h2>${isSignup ? "🎓 Create your student profile" : "👋 Welcome back"}</h2>
  <div class="auth-mode-toggle" role="tablist" aria-label="Account mode">
    <button type="button" role="tab" class="auth-mode-btn ${isSignup ? "active" : ""}" aria-selected="${isSignup}" onclick="switchStudentAuthMode('signup')">Sign up</button>
    <button type="button" role="tab" class="auth-mode-btn ${isSignup ? "" : "active"}" aria-selected="${!isSignup}" onclick="switchStudentAuthMode('signin')">Sign in</button>
  </div>
  <p class="auth-hint">${isSignup
      ? "No email, no password — your first name, last name and a number between 1 and 100 are your login."
      : "Enter the name and number you signed up with."}</p>
  <label for="stFirst">First name</label>
  <input type="text" id="stFirst" autocomplete="given-name" placeholder="ziad" value="${esc(first)}"/>
  <label for="stLast">Last name</label>
  <input type="text" id="stLast" autocomplete="family-name" placeholder="baroudi" value="${esc(last)}"/>
  <label for="stNumber">Your number (1–100)</label>
  <input type="number" id="stNumber" min="1" max="100" step="1" value="${esc(number)}"/>
  <div class="err" id="stAuthErr">${error ? esc(error) : ""}</div>
  <div class="modal-actions">
    <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
    <button class="btn btn-primary" onclick="submitStudentAuth('${isSignup ? "signup" : "signin"}')">${isSignup ? "Create profile" : "Sign in"}</button>
  </div>`);
}

function _readAuthValues() {
  return {
    first_name: document.getElementById("stFirst")?.value.trim() || "",
    last_name: document.getElementById("stLast")?.value.trim() || "",
    number: document.getElementById("stNumber")?.value.trim() || "",
  };
}

function _setAuthError(message) {
  const el = document.getElementById("stAuthErr");
  if (el) el.textContent = message || "";
}

function switchStudentAuthMode(mode) {
  _renderAuthModal({ mode, values: _readAuthValues() });
}

async function submitStudentAuth(mode) {
  const isSignup = mode !== "signin";
  const values = _readAuthValues();

  if (!values.first_name || !values.last_name) {
    _setAuthError("Please enter both your first and last name.");
    return;
  }
  const number = parseInt(values.number, 10);
  if (!Number.isInteger(number) || number < 1 || number > 100) {
    _setAuthError("Pick a whole number between 1 and 100.");
    return;
  }

  const btn = document.querySelector("#modalBox .btn-primary");
  setBtnLoading(btn, true, isSignup ? "Creating…" : "Signing in…");
  try {
    // The guest token rides along automatically: register claims that row,
    // login reassigns its page views onto the existing student.
    const data = await apiRequest(
      isSignup ? "/api/users/register" : "/api/users/login",
      {
        method: "POST",
        body: {
          first_name: values.first_name,
          last_name: values.last_name,
          number,
        },
      },
    );
    if (!data?.token) throw new Error("Sign-in response is missing a token");

    _setStudentToken(data.token);
    if (data.user) applyStudentUser(data.user);
    else await refreshStudentProfile();

    closeModal();
    const handle = studentHandle();
    showToast(
      isSignup
        ? `Profile created — welcome, ${handle || "student"}!`
        : `Signed in as ${handle || "student"}`,
    );

    const action = _pendingAction;
    _pendingAction = null;
    if (typeof action === "function") action();
  } catch (err) {
    if (isSignup && err?.status === 409) {
      const suggestion = _suggestNumber(number);
      _renderAuthModal({
        mode: "signup",
        values: { ...values, number: String(suggestion) },
        error: `${values.first_name} ${values.last_name} ${number} is already taken — that name + number pair must be unique. Try another number, e.g. ${suggestion}.`,
      });
      return;
    }
    if (!isSignup && err?.status === 404) {
      _renderAuthModal({
        mode: "signup",
        values,
        error: "No student with that name and number yet. Sign up instead — your details are already filled in.",
      });
      return;
    }
    if (err?.status === 400) {
      _setAuthError(formatApiError(err, "Please check your details and try again."));
      return;
    }
    logApiError(err, isSignup ? "studentRegister" : "studentLogin");
    _setAuthError(formatApiError(err, "Something went wrong. Please try again."));
  } finally {
    setBtnLoading(btn, false);
  }
}

async function signOutStudent() {
  await resetToGuest();
  if (AppState.currentProg === "favorites") window.selectProg?.("all");
  showToast("Signed out.");
  if (AppState.studentToken) {
    await refreshStudentProfile().catch((err) =>
      logApiError(err, "studentSession"),
    );
  }
}

// ── Welcome banner ──────────────────────────────────────────────────────────
function renderStudentBanner() {
  const el = document.getElementById("studentWelcome");
  if (!el) return;
  const handle = studentHandle();
  if (!handle && AppState.adminLoggedIn) {
    el.hidden = true;
    el.innerHTML = "";
    return;
  }
  el.innerHTML = handle
    ? `<span class="student-welcome-text">👋 Welcome, <strong>${esc(handle)}</strong></span>
       <button type="button" class="student-welcome-btn" data-action="studentSignOut">Sign out</button>`
    : `<span class="student-welcome-text">Browsing as a guest — sign up to open links, report issues and save courses.</span>
       <button type="button" class="student-welcome-btn" data-action="studentSignIn">Sign up / Sign in</button>`;
  el.hidden = false;
}

// ── Favorites ───────────────────────────────────────────────────────────────
/** Sync a single toggle. Throws so callers can roll back their optimistic UI. */
async function syncFavorite(courseId, added) {
  await apiRequest(`/api/users/me/favorites/${encodeURIComponent(courseId)}`, {
    method: added ? "POST" : "DELETE",
  });
  const user = AppState.studentUser;
  if (user) {
    const ids = new Set((user.favorite_course_ids || []).map(Number));
    if (added) ids.add(Number(courseId));
    else ids.delete(Number(courseId));
    user.favorite_course_ids = [...ids];
  }
}

/** Re-sync star buttons already on screen with AppState.favorites. */
function repaintFavoriteStars() {
  document.querySelectorAll(".course-card").forEach((card) => {
    const btn = card.querySelector(".fav-btn");
    if (!btn) return;
    const id = card.id.replace("course-card-", "");
    const isFav = AppState.favorites.has(String(id));
    btn.classList.toggle("active", isFav);
    btn.title = isFav ? "Remove from My Courses" : "Add to My Courses";
  });
  if (AppState.currentProg === "favorites") window.renderCourses?.();
}

Object.assign(window, {
  bootstrapStudentSession,
  requireStudent,
  promptStudentAuth,
  handleStudentAuthError,
  onStudentTokenRejected,
  submitStudentAuth,
  switchStudentAuthMode,
  signOutStudent,
  syncFavorite,
  studentHandle,
  isRegisteredStudent,
  renderStudentBanner,
  repaintFavoriteStars,
});

export {
  bootstrapStudentSession,
  requireStudent,
  promptStudentAuth,
  handleStudentAuthError,
  signOutStudent,
  syncFavorite,
  studentHandle,
  isRegisteredStudent,
  renderStudentBanner,
  repaintFavoriteStars,
};
