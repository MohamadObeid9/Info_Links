// ===================== API CORE =====================
import { AppState } from "./state.js";

const API_TIMEOUT_MS = 12000;

function _appendRequestRef(message, status, requestId) {
  if (
    status >= 500 &&
    typeof requestId === "string" &&
    requestId.trim() &&
    !message.includes("(ref:")
  ) {
    return `${message} (ref: ${requestId.trim()})`;
  }
  return message;
}

function _buildApiError(status, fallbackMessage, payloadText) {
  if (!payloadText) return new Error(fallbackMessage);
  try {
    const parsed = JSON.parse(payloadText);
    if (parsed && typeof parsed.error === "string" && parsed.error.trim()) {
      const message = _appendRequestRef(
        parsed.error,
        status,
        parsed.request_id,
      );
      const err = new Error(message);
      if (typeof parsed.request_id === "string" && parsed.request_id.trim()) {
        err.requestId = parsed.request_id.trim();
      }
      if (typeof parsed.code === "string" && parsed.code.trim()) {
        err.code = parsed.code.trim();
      }
      return err;
    }
  } catch (e) {}
  return new Error(
    payloadText || fallbackMessage || `Request failed (${status})`,
  );
}

/** User-facing text from an API error; falls back when message is empty. */
function formatApiError(err, fallback = "Something went wrong") {
  if (err && typeof err.message === "string" && err.message.trim()) {
    return err.message;
  }
  return fallback;
}

/** Structured console log for API failures (message + request id when present). */
function logApiError(err, context, status) {
  const entry = {
    context: context || "api",
    message: err?.message || String(err),
  };
  if (status) entry.status = status;
  if (err?.requestId) entry.requestId = err.requestId;
  console.error("[API error]", entry);
}

// Endpoints where the server derives the acting student from the token, so the
// student session token wins over any admin token present in the same browser.
const STUDENT_AUTH_PATHS =
  /^\/api\/(page_views|link_clicks|service_clicks|search_events|reports|feedback|contributions|users)(\/|$|\?)/;

function _usesStudentToken(url) {
  return STUDENT_AUTH_PATHS.test(String(url).split("?")[0]);
}

async function apiRequest(
  url,
  {
    method = "GET",
    body = null,
    headers = {},
    timeoutMs = API_TIMEOUT_MS,
    cache,
  } = {},
) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  const finalHeaders = { ...headers };
  if (body !== null && !finalHeaders["Content-Type"]) {
    finalHeaders["Content-Type"] = "application/json";
  }
  if (!finalHeaders.Authorization) {
    if (AppState.studentToken && _usesStudentToken(url)) {
      finalHeaders.Authorization = `Bearer ${AppState.studentToken}`;
    } else if (AppState.sbToken) {
      finalHeaders.Authorization = `Bearer ${AppState.sbToken}`;
    }
  }

  try {
    const res = await fetch(url, {
      method,
      headers: finalHeaders,
      body: body !== null ? JSON.stringify(body) : null,
      signal: controller.signal,
      ...(cache ? { cache } : {}),
    });
    const text = await res.text();
    if (!res.ok) {
      let apiErr = _buildApiError(
        res.status,
        `Request failed (${res.status})`,
        text,
      );
      apiErr.status = res.status;
      const headerId = res.headers.get("X-Request-ID");
      if (!apiErr.requestId && headerId?.trim()) {
        apiErr.requestId = headerId.trim();
        if (res.status >= 500 && !apiErr.message.includes("(ref:")) {
          apiErr.message = _appendRequestRef(
            apiErr.message,
            res.status,
            headerId,
          );
        }
      }
      logApiError(apiErr, `${method} ${url}`, res.status);
      throw apiErr;
    }
    return text ? JSON.parse(text) : [];
  } catch (err) {
    if (err && err.name === "AbortError") {
      throw new Error("Request timed out. Please try again.");
    }
    throw err;
  } finally {
    clearTimeout(timeoutId);
  }
}

// ===================== API PROXY =====================
async function sb(
  table,
  method = "GET",
  body = null,
  matchString = null,
  _select = null,
) {
  let cleanTable = table;
  let id = null;

  if (table.includes("?id=eq.")) {
    const parts = table.split("?id=eq.");
    cleanTable = parts[0];
    id = parts[1];
  } else if (matchString && matchString.includes("id=eq.")) {
    id = matchString.split("id=eq.")[1];
  }

  let url = `/api/admin/${cleanTable}`;
  if (id) url += `/${id}`;

  return apiRequest(url, { method, body });
}

async function sbAuth(email, password) {
  const data = await apiRequest("/api/auth/login", {
    method: "POST",
    body: { email, password },
  });
  return data.token;
}

async function sbLogout() {
  AppState.sbToken = null;
  localStorage.removeItem("infolinks_token");
}

function _visitDay() {
  const d = new Date();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${m}-${day}`;
}

function _visitGuardKey(uid) {
  return `${uid == null ? "anon" : uid}:${_visitDay()}`;
}

let _visitInflight = null;

async function trackVisit() {
  if (AppState.adminLoggedIn) return;
  if (!AppState.studentToken) return;
  const uid = AppState.studentUser?.id;
  const key = _visitGuardKey(uid);
  let tracked = "";
  try {
    tracked = sessionStorage.getItem("pv_tracked") || "";
  } catch (e) { /* private mode */ }
  // Once per user per local calendar day. A tab left open overnight, or restored
  // by the browser, still records the next day. Extra rows the same day collapse
  // in analytics (one person per day).
  if (tracked === key) return;
  if (_visitInflight) return _visitInflight;
  _visitInflight = (async () => {
    try {
      await apiRequest(`/api/page_views`, {
        method: "POST",
        body: { page: "home" },
      });
      try {
        sessionStorage.setItem("pv_tracked", key);
      } catch (e) { /* private mode */ }
    } catch (e) {
      if (e?.status === 401) window.onStudentTokenRejected?.();
    } finally {
      _visitInflight = null;
    }
  })();
  return _visitInflight;
}

function markVisitRecordedToday() {
  const uid = AppState.studentUser?.id;
  if (uid == null) return;
  try {
    sessionStorage.setItem("pv_tracked", _visitGuardKey(uid));
  } catch (e) { /* private mode */ }
}

function trackLinkClick(linkId, linkKind = "link", programId = null) {
  if (!linkId || AppState.adminLoggedIn) return;
  // A click means they are here today, even if this tab already counted an older visit.
  trackVisit();
  const payload =
    linkKind === "extra_link"
      ? { extra_link_id: linkId }
      : { link_id: linkId };
  const pid = Number(programId);
  if (Number.isFinite(pid) && pid > 0) payload.program_id = pid;
  apiRequest(`/api/link_clicks`, {
    method: "POST",
    body: payload,
  }).catch((e) => {
    if (e?.status === 401) window.onStudentTokenRejected?.();
  });
}

let _searchTrackTimer = null;
let _lastSearchTracked = "";

function trackSearch(query) {
  if (AppState.adminLoggedIn || !AppState.studentToken) return;
  const q = String(query || "").trim().toLowerCase();
  if (q.length < 2 || q === _lastSearchTracked) return;
  trackVisit();
  clearTimeout(_searchTrackTimer);
  _searchTrackTimer = setTimeout(() => {
    flushSearch(q);
  }, 1200);
}

function flushSearch(queryOverride) {
  clearTimeout(_searchTrackTimer);
  _searchTrackTimer = null;
  if (AppState.adminLoggedIn || !AppState.studentToken) return;
  const input = document.getElementById("searchInput");
  const q = (typeof queryOverride === "string" ? queryOverride : (input?.value || "")).trim().toLowerCase();
  if (q.length < 2 || q === _lastSearchTracked) return;
  _lastSearchTracked = q;
  apiRequest(`/api/search_events`, {
    method: "POST",
    body: { query: q },
  }).catch((e) => {
    if (e?.status === 401) window.onStudentTokenRejected?.();
  });
}

// Global Bridge
window.sb = sb;
window.sbAuth = sbAuth;
window.sbLogout = sbLogout;
window.trackVisit = trackVisit;
window.markVisitRecordedToday = markVisitRecordedToday;
window.trackLinkClick = trackLinkClick;
window.trackSearch = trackSearch;
window.flushSearch = flushSearch;
window.apiRequest = apiRequest;
window.formatApiError = formatApiError;
window.logApiError = logApiError;

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible") trackVisit();
});
window.addEventListener("pageshow", () => {
  trackVisit();
});

export {
  sb,
  sbAuth,
  sbLogout,
  trackVisit,
  trackLinkClick,
  trackSearch,
  flushSearch,
  apiRequest,
  formatApiError,
  logApiError,
};
