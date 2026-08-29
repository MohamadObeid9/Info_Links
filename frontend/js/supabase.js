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
  /^\/api\/(page_views|link_clicks|service_clicks|search_events|browse_events|reports|feedback|contributions|users)(\/|$|\?)/;

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
  select = null,
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

async function trackVisit() {
  if (AppState.adminLoggedIn) return;
  if (!AppState.studentToken) return;
  const uid = AppState.studentUser?.id;
  const tracked = sessionStorage.getItem("pv_tracked");
  // Guard is per user id so a re-bootstrapped guest still gets a page view.
  if (uid != null && tracked === String(uid)) return;
  if (tracked === "1" && uid == null) return;
  if (tracked === "1" && uid != null) {
    sessionStorage.setItem("pv_tracked", String(uid));
    return;
  }
  try {
    await apiRequest(`/api/page_views`, {
      method: "POST",
      body: { page: "home" },
    });
    sessionStorage.setItem("pv_tracked", uid != null ? String(uid) : "1");
  } catch (e) {
    if (e?.status === 401) window.onStudentTokenRejected?.();
  }
}

function trackLinkClick(linkId, linkKind = "link") {
  if (!linkId || AppState.adminLoggedIn) return;
  const payload =
    linkKind === "extra_link"
      ? { extra_link_id: linkId }
      : { link_id: linkId };
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
  clearTimeout(_searchTrackTimer);
  _searchTrackTimer = setTimeout(() => {
    _lastSearchTracked = q;
    apiRequest(`/api/search_events`, {
      method: "POST",
      body: { query: q },
    }).catch((e) => {
      if (e?.status === 401) window.onStudentTokenRejected?.();
    });
  }, 600);
}

function trackBrowse(step) {
  if (AppState.adminLoggedIn || !AppState.studentToken) return;
  if (step !== "year" && step !== "list") return;
  const key = `browse_${step}`;
  if (sessionStorage.getItem(key)) return;
  sessionStorage.setItem(key, "1");
  apiRequest(`/api/browse_events`, {
    method: "POST",
    body: { step },
  }).catch((e) => {
    if (e?.status === 401) {
      sessionStorage.removeItem(key);
      window.onStudentTokenRejected?.();
    }
  });
}

// Global Bridge
window.sb = sb;
window.sbAuth = sbAuth;
window.sbLogout = sbLogout;
window.trackVisit = trackVisit;
window.trackLinkClick = trackLinkClick;
window.trackSearch = trackSearch;
window.trackBrowse = trackBrowse;
window.apiRequest = apiRequest;
window.formatApiError = formatApiError;
window.logApiError = logApiError;

export {
  sb,
  sbAuth,
  sbLogout,
  trackVisit,
  trackLinkClick,
  trackSearch,
  trackBrowse,
  apiRequest,
  formatApiError,
  logApiError,
};
