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

/** Clear once-per-tab analytics guards so a new guest identity can record activity. */
function _clearSessionAnalyticsGuards() {
  try {
    sessionStorage.removeItem("pv_tracked");
    sessionStorage.removeItem("browse_year");
    sessionStorage.removeItem("browse_list");
  } catch (e) { }
}

function applyStudentUser(user) {
  const previousId = AppState.studentUser?.id ?? AppState.studentUserId;
  if (previousId && user?.id && previousId !== user.id) _forgetStudentIdentity();

  AppState.studentUser = user || null;
  if (user && !user.is_guest) {
    _rememberAccount();
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
  _clearSessionAnalyticsGuards();
  renderStudentBanner();
  repaintFavoriteStars();
  try {
    await createGuestSession();
    // New guest id — record a visit (the old pv_tracked guard would skip this).
    await window.trackVisit?.();
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
      // Bind or record the visit now that studentUser.id is known.
      await window.trackVisit?.();
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
const HAS_ACCOUNT_KEY = "infolinks_has_account";

// Credentials from a signup that hit an existing name, kept for "that's me".
let _pendingSignup = null;
// Runs after the auth modal closes following a successful sign-in or the
// "account created" screen. Opening a link waits until that screen is dismissed.
let _afterAuthClose = null;

function _authIcon(paths) {
  return `<svg viewBox="0 0 24 24" fill="none" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${paths}</svg>`;
}

const AUTH_ICON_USER_CHECK = _authIcon(
  '<path d="M8 7a4 4 0 1 0 8 0a4 4 0 0 0-8 0"/><path d="M6 21v-2a4 4 0 0 1 4-4h2"/><path d="m15 19 2 2 4-4"/>',
);
const AUTH_ICON_USER_PLUS = _authIcon(
  '<path d="M8 7a4 4 0 1 0 8 0a4 4 0 0 0-8 0"/><path d="M6 21v-2a4 4 0 0 1 4-4h4"/><path d="M16 19h6M19 16v6"/>',
);
const AUTH_ICON_USER = _authIcon(
  '<path d="M8 7a4 4 0 1 0 8 0a4 4 0 0 0-8 0"/><path d="M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/>',
);
const AUTH_ICON_PHONE = _authIcon(
  '<rect x="7" y="4" width="10" height="16" rx="2"/><path d="M11 17h2"/>',
);
const AUTH_ICON_LAPTOP = _authIcon(
  '<rect x="3" y="5" width="18" height="12" rx="2"/><path d="M2 20h20"/>',
);

/** Next number to suggest after a collision (55 → 65, wrapping past 100). */
function _suggestNumber(taken) {
  const next = Number(taken) + 10;
  return next > 100 ? next - 100 : next;
}

function _rememberAccount() {
  try {
    localStorage.setItem(HAS_ACCOUNT_KEY, "1");
  } catch (e) { }
}

function _hasAccountOnDevice() {
  try {
    return localStorage.getItem(HAS_ACCOUNT_KEY) === "1";
  } catch (e) {
    return false;
  }
}

function _initialAuthScreen(mode) {
  if (mode === "signup" || mode === "signin" || mode === "choose") return mode;
  return _hasAccountOnDevice() ? "signin" : "choose";
}

function promptStudentAuth({ retry = null, mode } = {}) {
  _pendingAction = typeof retry === "function" ? retry : null;
  _pendingSignup = null;
  _afterAuthClose = null;
  _renderAuthModal(_initialAuthScreen(mode));
}

function _renderAuthModal(screen) {
  openModal(`<div class="auth-shell">
    <section data-screen="choose">
      <h2 id="authTitleChoose">Welcome to Info Links</h2>
      <p class="auth-sub">Have you used Info Links before, on any device?</p>
      <button class="auth-choice auth-choice-main" data-auth-go="signin" type="button">
        ${AUTH_ICON_USER_CHECK}
        <div><b>Yes, I have an account</b><span>Sign in with your name and number</span></div>
      </button>
      <button class="auth-choice" data-auth-go="signup" type="button">
        ${AUTH_ICON_USER_PLUS}
        <div><b>No, I'm new</b><span>Create a free account</span></div>
      </button>
      <div class="auth-same">
        ${AUTH_ICON_PHONE}
        <span aria-hidden="true">=</span>
        ${AUTH_ICON_LAPTOP}
        <span>One login for all your devices</span>
      </div>
      <button type="button" class="auth-close" data-auth-action="cancel" aria-label="Close">×</button>
    </section>

    <section data-screen="signin" hidden>
      <h2 id="authTitleSignin">Welcome back</h2>
      <p class="auth-sub">Use the same name and number you chose the first time, even if that was on your other device.</p>
      <div class="auth-field"><label for="siFirst">First name</label><input id="siFirst" autocomplete="off" placeholder="ziad"></div>
      <div class="auth-field"><label for="siLast">Last name</label><input id="siLast" autocomplete="off" placeholder="baroudi"></div>
      <div class="auth-field"><label for="siPin">Your secret number (1–100)</label><input id="siPin" inputmode="numeric" autocomplete="off" placeholder="33"></div>
      <p class="auth-err" id="siErr" role="alert"></p>
      <button class="btn btn-primary" id="authSignInBtn" data-auth-action="signin" type="button">Sign in</button>
      <div class="auth-row">
        <button class="btn btn-ghost" data-auth-go="choose" type="button">Back</button>
        <button class="btn btn-ghost" data-auth-action="cancel" type="button">Cancel</button>
      </div>
      <p class="auth-foot">First time here? <button class="auth-link" data-auth-go="signup" type="button">Create an account</button></p>
    </section>

    <section data-screen="signup" hidden>
      <h2 id="authTitleSignup">Create your account</h2>
      <div class="auth-note auth-note-warn"><b>Already signed up on your phone or laptop? Don't create a new account.</b> <button class="auth-link" data-auth-go="signin" type="button">Sign in instead</button>. Extra accounts are deleted.</div>
      <div class="auth-field"><label for="suFirst">First name</label><input id="suFirst" autocomplete="off" placeholder="ziad"></div>
      <div class="auth-field"><label for="suLast">Last name</label><input id="suLast" autocomplete="off" placeholder="baroudi"></div>
      <div class="auth-field"><label for="suPin">Pick a secret number (1–100)</label><input id="suPin" inputmode="numeric" autocomplete="off" placeholder="33"></div>
      <p class="auth-pin-help">This is not your phone number. You'll need it to sign in on your other device.</p>
      <p class="auth-err" id="suErr" role="alert"></p>
      <button class="btn btn-primary" id="authSignUpBtn" data-auth-action="signup" type="button">Create account</button>
      <div class="auth-row">
        <button class="btn btn-ghost" data-auth-go="choose" type="button">Back</button>
        <button class="btn btn-ghost" data-auth-action="cancel" type="button">Cancel</button>
      </div>
    </section>

    <section data-screen="duplicate" hidden>
      <h2 id="authTitleDup">This name already exists</h2>
      <div class="auth-who">
        ${AUTH_ICON_USER}
        <div><b id="dupName"></b><span>Account created on another device</span></div>
      </div>
      <p class="auth-sub">Is that you? Then sign in instead of making a second account. Two accounts for one person get deleted.</p>
      <button class="btn btn-primary" id="authDupYes" data-auth-action="dup-yes" type="button">Yes, that's me. Sign me in</button>
      <button class="btn btn-ghost" id="authDupNo" data-auth-action="dup-no" type="button">No, I'm a different person</button>
    </section>

    <section data-screen="saved" hidden>
      <h2 id="authTitleSaved">Account created</h2>
      <p class="auth-sub">Save your login. You'll need it on your other device.</p>
      <div class="auth-login"><small>Your login</small><strong id="authLoginText"></strong></div>
      <div class="auth-row auth-row-gap">
        <button class="btn btn-ghost" id="authCopyBtn" data-auth-action="copy" type="button">Copy</button>
        <button class="btn btn-ghost" data-auth-action="save-image" type="button">Save as image</button>
      </div>
      <div class="auth-note auth-note-info">On your other device, tap <b>Sign in</b> and enter exactly this. Don't sign up again.</div>
      <button class="btn btn-primary" data-auth-action="done" type="button">Got it, continue</button>
    </section>
  </div>`);
  _showAuthScreen(screen);
}

function _showAuthScreen(name) {
  const shell = document.querySelector(".auth-shell");
  if (!shell) return;
  shell.querySelectorAll("[data-screen]").forEach((section) => {
    section.hidden = section.dataset.screen !== name;
  });
  const heading = shell.querySelector("[data-screen]:not([hidden]) h2");
  if (heading?.id) {
    document.getElementById("modalBox")?.setAttribute("aria-labelledby", heading.id);
  }
  const visible = "[data-screen]:not([hidden])";
  const first = shell.querySelector(
    `${visible} input, ${visible} .auth-choice, ${visible} .btn`,
  );
  if (first && typeof first.focus === "function") first.focus();
}

function _readAuth(prefix) {
  return {
    first_name: document.getElementById(prefix + "First")?.value.trim() || "",
    last_name: document.getElementById(prefix + "Last")?.value.trim() || "",
    pinRaw: document.getElementById(prefix + "Pin")?.value.trim() || "",
  };
}

function _setAuthError(id, message) {
  const el = document.getElementById(id);
  if (el) el.textContent = message || "";
}

function _validateAuth(prefix, errId) {
  const values = _readAuth(prefix);
  if (!values.first_name || !values.last_name) {
    _setAuthError(errId, "Enter your first and last name.");
    return null;
  }
  const number = Number(values.pinRaw);
  if (!/^\d+$/.test(values.pinRaw) || number < 1 || number > 100) {
    _setAuthError(errId, "Your secret number must be between 1 and 100. It is not your phone number.");
    return null;
  }
  _setAuthError(errId, "");
  return { first_name: values.first_name, last_name: values.last_name, number };
}

function _authFailMessage(err) {
  if (err?.status === 429) return "Too many tries. Wait a few minutes and try again.";
  if (!err?.status) return "Couldn't reach Info Links. Check your connection and try again.";
  return formatApiError(err, "Something went wrong. Please try again.");
}

function _copySignupIntoSignIn() {
  const signup = _readAuth("su");
  const first = document.getElementById("siFirst");
  const last = document.getElementById("siLast");
  if (first && !first.value && signup.first_name) first.value = signup.first_name;
  if (last && !last.value && signup.last_name) last.value = signup.last_name;
}

function _copySignInIntoSignup() {
  const signin = _readAuth("si");
  const first = document.getElementById("suFirst");
  const last = document.getElementById("suLast");
  if (first && !first.value && signin.first_name) first.value = signin.first_name;
  if (last && !last.value && signin.last_name) last.value = signin.last_name;
}

function _armAuthClose() {
  const action = _pendingAction;
  _pendingAction = null;
  _afterAuthClose = typeof action === "function" ? action : null;
}

async function _applyAuthSession(data) {
  if (!data?.token) throw new Error("Sign-in response is missing a token");
  _setStudentToken(data.token);
  if (data.user) applyStudentUser(data.user);
  else await refreshStudentProfile();
  _rememberAccount();
}

function _showSavedLogin(creds) {
  const text = `${creds.first_name} · ${creds.last_name} · ${creds.number}`;
  const label = document.getElementById("authLoginText");
  if (label) label.textContent = text;
  const copyBtn = document.getElementById("authCopyBtn");
  if (copyBtn) copyBtn.textContent = "Copy";
  _showAuthScreen("saved");
}

async function _submitSignIn() {
  const creds = _validateAuth("si", "siErr");
  if (!creds) return;
  const btn = document.getElementById("authSignInBtn");
  setBtnLoading(btn, true, "Signing in…");
  try {
    const data = await apiRequest("/api/users/login", { method: "POST", body: creds });
    await _applyAuthSession(data);
    window.markVisitRecordedToday?.();
    const handle = studentHandle();
    showToast(`Signed in as ${handle || "student"}`);
    _armAuthClose();
    closeModal();
  } catch (err) {
    if (err?.status === 404) {
      _setAuthError("siErr", "That name and number don't match. Check the spelling and the number you picked when you signed up.");
      return;
    }
    if (err?.status !== 400 && err?.status !== 429) logApiError(err, "studentLogin");
    _setAuthError("siErr", err?.status === 400
      ? formatApiError(err, "Please check your details and try again.")
      : _authFailMessage(err));
  } finally {
    setBtnLoading(btn, false);
  }
}

async function _submitSignUp(confirmDifferent) {
  const creds = confirmDifferent ? _pendingSignup : _validateAuth("su", "suErr");
  if (!creds) return;
  _pendingSignup = creds;
  const btn = document.getElementById(confirmDifferent ? "authDupNo" : "authSignUpBtn");
  setBtnLoading(btn, true, "Creating…");
  try {
    const data = await apiRequest("/api/users/register", {
      method: "POST",
      body: { ...creds, confirm_different: !!confirmDifferent },
    });
    await _applyAuthSession(data);
    // Claim keeps the same id (visit already recorded). Fallthrough create is a
    // new id with no page_views yet — trackVisit records one when needed.
    try {
      await window.trackVisit?.();
    } catch (visitErr) {
      logApiError(visitErr, "studentRegister");
    }
    _armAuthClose();
    _showSavedLogin(creds);
  } catch (err) {
    if (err?.status === 409 && err?.code === "name_exists") {
      const name = document.getElementById("dupName");
      if (name) name.textContent = `${creds.first_name} ${creds.last_name}`;
      _showAuthScreen("duplicate");
      return;
    }
    if (err?.status === 409) {
      const suggestion = _suggestNumber(creds.number);
      _showAuthScreen("signup");
      const pin = document.getElementById("suPin");
      if (pin) pin.value = String(suggestion);
      _setAuthError("suErr", `${creds.first_name} ${creds.last_name} #${creds.number} is already taken. Try another number, for example ${suggestion}.`);
      return;
    }
    _showAuthScreen("signup");
    if (err?.status !== 400 && err?.status !== 429) logApiError(err, "studentRegister");
    _setAuthError("suErr", err?.status === 400
      ? formatApiError(err, "Please check your details and try again.")
      : _authFailMessage(err));
  } finally {
    setBtnLoading(btn, false);
  }
}

function _signInAsPending() {
  _showAuthScreen("signin");
  if (!_pendingSignup) return;
  const first = document.getElementById("siFirst");
  const last = document.getElementById("siLast");
  const pin = document.getElementById("siPin");
  if (first) first.value = _pendingSignup.first_name;
  if (last) last.value = _pendingSignup.last_name;
  if (pin) {
    pin.value = "";
    pin.focus();
  }
}

function _copyTextFallback(text) {
  const area = document.createElement("textarea");
  area.value = text;
  area.setAttribute("aria-hidden", "true");
  area.style.position = "fixed";
  area.style.top = "0";
  area.style.left = "0";
  area.style.width = "2em";
  area.style.height = "2em";
  area.style.padding = "0";
  area.style.border = "none";
  area.style.outline = "none";
  area.style.opacity = "0";
  document.body.appendChild(area);
  area.focus();
  area.select();
  area.setSelectionRange(0, text.length);
  let ok = false;
  try {
    ok = document.execCommand("copy");
  } catch (e) {
    ok = false;
  }
  area.remove();
  return ok;
}

function _selectLoginText() {
  const el = document.getElementById("authLoginText");
  if (!el) return;
  const range = document.createRange();
  range.selectNodeContents(el);
  const selection = window.getSelection();
  selection?.removeAllRanges();
  selection?.addRange(range);
}

async function _copySavedLogin() {
  const text = document.getElementById("authLoginText")?.textContent || "";
  const btn = document.getElementById("authCopyBtn");
  // execCommand only succeeds during this click, so try it before any await.
  let copied = _copyTextFallback(text);
  if (!copied && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
    } catch (e) {
      copied = false;
    }
  }
  if (copied) {
    if (btn) btn.textContent = "Copied";
    return;
  }
  _selectLoginText();
  if (btn) btn.textContent = "Press Ctrl+C";
}

function _saveLoginImage() {
  const text = document.getElementById("authLoginText")?.textContent || "";
  const canvas = document.createElement("canvas");
  canvas.width = 900;
  canvas.height = 420;
  const g = canvas.getContext("2d");
  if (!g) return;
  g.fillStyle = "#ffffff";
  g.fillRect(0, 0, 900, 420);
  g.fillStyle = "#6c63ff";
  g.fillRect(0, 0, 900, 16);
  g.fillStyle = "#17172b";
  g.font = "600 34px Inter, system-ui, sans-serif";
  g.fillText("Info Links login", 50, 90);
  let size = 42;
  g.font = `600 ${size}px Inter, system-ui, sans-serif`;
  while (size > 18 && g.measureText(text).width > 800) {
    size -= 2;
    g.font = `600 ${size}px Inter, system-ui, sans-serif`;
  }
  g.fillText(text, 50, 210);
  g.fillStyle = "#6b6b85";
  g.font = "26px Inter, system-ui, sans-serif";
  g.fillText("On your other device: tap Sign in and enter exactly this.", 50, 290);
  g.fillText("Do not sign up again.", 50, 330);
  const link = document.createElement("a");
  link.download = "info-links-login.png";
  link.href = canvas.toDataURL("image/png");
  link.click();
}

function switchStudentAuthMode(mode) {
  if (!document.querySelector(".auth-shell")) {
    promptStudentAuth({ mode });
    return;
  }
  if (mode === "signin") _copySignupIntoSignIn();
  if (mode === "signup") _copySignInIntoSignup();
  _showAuthScreen(mode === "signup" ? "signup" : "signin");
}

document.addEventListener("click", (e) => {
  const shell = e.target.closest?.(".auth-shell");
  if (!shell) return;
  const go = e.target.closest("[data-auth-go]");
  if (go) {
    e.preventDefault();
    if (go.dataset.authGo === "signin") _copySignupIntoSignIn();
    if (go.dataset.authGo === "signup") _copySignInIntoSignup();
    _showAuthScreen(go.dataset.authGo);
    return;
  }
  const action = e.target.closest("[data-auth-action]")?.dataset.authAction;
  if (!action) return;
  e.preventDefault();
  if (action === "cancel" || action === "done") closeModal();
  else if (action === "signin") _submitSignIn();
  else if (action === "signup") _submitSignUp(false);
  else if (action === "dup-yes") _signInAsPending();
  else if (action === "dup-no") _submitSignUp(true);
  else if (action === "copy") _copySavedLogin();
  else if (action === "save-image") _saveLoginImage();
});

document.addEventListener("keydown", (e) => {
  if (e.key !== "Enter" || e.target.tagName !== "INPUT") return;
  const screen = e.target.closest(".auth-shell [data-screen]")?.dataset.screen;
  if (screen !== "signin" && screen !== "signup") return;
  e.preventDefault();
  if (screen === "signin") _submitSignIn();
  else _submitSignUp(false);
});

const _authModalEl = document.getElementById("modal");
if (_authModalEl) {
  new MutationObserver(() => {
    if (_authModalEl.classList.contains("open")) return;
    const action = _afterAuthClose;
    _afterAuthClose = null;
    if (typeof action === "function") action();
  }).observe(_authModalEl, { attributes: true, attributeFilter: ["class"] });
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

function showStudentProfileModal() {
  const u = AppState.studentUser;
  if (!u || u.is_guest) return;
  const first = u.first_name || "";
  const last = u.last_name || "";
  const num = u.number || "";
  openModal(`<h2>📱💻 Your Student Login Info</h2>
  <p class="auth-hint">Use these exact details to sign in on your phone or laptop so your saved courses and history stay in sync:</p>
  <div class="student-credentials-card">
    <div class="cred-row"><span class="cred-label">First name:</span> <strong class="cred-val">${esc(first)}</strong></div>
    <div class="cred-row"><span class="cred-label">Last name:</span> <strong class="cred-val">${esc(last)}</strong></div>
    <div class="cred-row"><span class="cred-label">Student number:</span> <strong class="cred-val cred-num">#${esc(String(num))}</strong></div>
  </div>
  <p class="auth-hint" style="margin-top:12px;"><strong>Note:</strong> Your student number is a number between 1 and 100 (this is a PIN, NOT your mobile phone number). Enter these 3 fields in <strong>Sign in</strong> on any device.</p>
  <div class="modal-actions">
    <button class="btn btn-primary" onclick="closeModal()">Got it</button>
  </div>`);
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
  const u = AppState.studentUser;
  if (u && !u.is_guest) {
    const displayName = `${u.first_name || ""} ${u.last_name || ""}`.trim() || handle;
    el.innerHTML = `<span class="student-welcome-text">👋 Welcome, <strong>${esc(displayName)}</strong> <button type="button" class="student-id-badge" data-action="studentProfileInfo" title="View your login credentials">#${esc(String(u.number))}</button></span>
       <button type="button" class="student-welcome-btn" data-action="studentProfileInfo">📱💻 Sync / Login info</button>
       <button type="button" class="student-welcome-btn" data-action="studentSignOut">Sign out</button>`;
  } else {
    el.innerHTML = `<span class="student-welcome-text">Browsing as a guest — sign in or sign up to open links, report issues and save courses.</span>
       <button type="button" class="student-welcome-btn" data-action="studentSignIn">Sign in / Sign up</button>`;
  }
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
    const id = card.dataset.courseId || card.id.replace(/^course-card-/, "").replace(/-p\d+$/, "");
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
  switchStudentAuthMode,
  signOutStudent,
  syncFavorite,
  studentHandle,
  isRegisteredStudent,
  renderStudentBanner,
  repaintFavoriteStars,
  showStudentProfileModal,
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
  showStudentProfileModal,
};
