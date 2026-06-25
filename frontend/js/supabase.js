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
  if (AppState.sbToken && !finalHeaders.Authorization) {
    finalHeaders.Authorization = `Bearer ${AppState.sbToken}`;
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
  if (sessionStorage.getItem("pv_tracked")) return;
  try {
    await apiRequest(`/api/page_views`, {
      method: "POST",
      body: { page: "home" },
    });
    sessionStorage.setItem("pv_tracked", "1");
  } catch (e) {}
}

function trackLinkClick(linkId) {
  if (!linkId || AppState.adminLoggedIn) return;
  apiRequest(`/api/link_clicks`, {
    method: "POST",
    body: { link_id: linkId },
  }).catch(() => {});
}

// Global Bridge
window.sb = sb;
window.sbAuth = sbAuth;
window.sbLogout = sbLogout;
window.trackVisit = trackVisit;
window.trackLinkClick = trackLinkClick;
window.apiRequest = apiRequest;
window.formatApiError = formatApiError;
window.logApiError = logApiError;

export {
  sb,
  sbAuth,
  sbLogout,
  trackVisit,
  trackLinkClick,
  apiRequest,
  formatApiError,
  logApiError,
};
