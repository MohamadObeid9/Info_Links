

// ===================== API PROXY =====================
async function sb(table, method = "GET", body = null, matchString = null, select = null) {
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

  const headers = {
    "Content-Type": "application/json",
  };
  if (AppState.sbToken) {
    headers.Authorization = `Bearer ${AppState.sbToken}`;
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
  });

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || "Backend API Error");
  }

  const text = await res.text();
  return text ? JSON.parse(text) : [];
}

async function sbAuth(email, password) {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error("Login failed");
  const data = await res.json();
  return data.token;
}

async function sbLogout() {
  AppState.sbToken = null;
  localStorage.removeItem("infolinks_token");
}

async function trackVisit() {
  if (sessionStorage.getItem("pv_tracked")) return;
  try {
    await fetch(`/api/page_views`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ page: "home" }),
    });
    sessionStorage.setItem("pv_tracked", "1");
  } catch (e) {}
}

function trackLinkClick(linkId) {
  if (!linkId) return;
  fetch(`/api/link_clicks`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ link_id: linkId }),
  }).catch(() => {});
}

// Global Bridge
window.sb = sb;
window.sbAuth = sbAuth;
window.sbLogout = sbLogout;
window.trackVisit = trackVisit;
window.trackLinkClick = trackLinkClick;
