import { AppState } from "./state.js";
import { esc, _linkHref, isMobileView, setSectionHint, COMMUNITY_PROMOTE_HINT, COMMUNITY_HINT_CARD, homeSectionHeading, sectionInlineHintHtml, setHomeLegendVisible } from "./ui.js";
import { apiRequest } from "./supabase.js";

// ===================== COMMUNITY SERVICES =====================

const SIDEBAR_SERVICE_COUNT = 5;

// New on every full page load so service picks change after reload.
const PAGE_LOAD_SALT = `${Date.now()}_${Math.random()}`;

function _deviceType() {
  return isMobileView() ? "phone" : "laptop";
}

function _hash(str) {
  let h = 0;
  for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0;
  return h;
}

function _seededShuffle(items, seed) {
  const arr = items.slice();
  let s = seed;
  for (let i = arr.length - 1; i > 0; i--) {
    s = (s * 1103515245 + 12345) & 0x7fffffff;
    const j = s % (i + 1);
    [arr[i], arr[j]] = [arr[j], arr[i]];
  }
  return arr;
}

function _contextSeed(context) {
  return _hash(`${context}_${PAGE_LOAD_SALT}`);
}

function _activeServices(excludeIds = []) {
  return AppState.dbServices.filter(
    (s) => s.status !== "frozen" && !excludeIds.includes(s.id),
  );
}

function _randomShuffle(items) {
  const arr = items.slice();
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [arr[i], arr[j]] = [arr[j], arr[i]];
  }
  return arr;
}

function shuffleServices(context) {
  const active = _activeServices();
  if (!active.length) return [];
  if (context && String(context).startsWith("community")) {
    return _randomShuffle(active);
  }
  return _seededShuffle(active, _contextSeed(context));
}

function pickServices(count, context, excludeIds = []) {
  const active = _activeServices(excludeIds);
  if (!active.length) return [];
  const shuffled = _seededShuffle(active, _contextSeed(context));
  const n = Number.isFinite(count) && count > 0 ? count : active.length;
  return shuffled.slice(0, Math.min(n, shuffled.length));
}

/** Last service shown in a semester course list; next pick avoids it when possible. */
let lastSemesterServiceId = null;

function pickRotatingService() {
  const active = _activeServices();
  if (!active.length) return null;
  const others = active.filter((s) => s.id !== lastSemesterServiceId);
  const pool = others.length ? others : active;
  const picked = pool[Math.floor(Math.random() * pool.length)];
  lastSemesterServiceId = picked.id;
  return picked;
}

function whatsappHref(phone) {
  const raw = String(phone || "").trim();
  if (!raw) return "";
  if (/^https?:\/\/(wa\.me|api\.whatsapp\.com)\//i.test(raw)) return raw;
  if (/^whatsapp:/i.test(raw)) {
    const match = raw.match(/phone=([0-9]+)/i);
    if (match) return `https://wa.me/${match[1]}`;
  }
  const digits = raw.replace(/\D/g, "");
  return digits ? `https://wa.me/${digits}` : "";
}

function _isWhatsAppUrl(parsed) {
  if (parsed.protocol === "whatsapp:") return true;
  const host = parsed.hostname.toLowerCase();
  return host === "wa.me" || host === "api.whatsapp.com";
}

function primaryServiceContact(s) {
  if (!s) return "";
  if (s.url) return s.url;
  if (s.phone) return whatsappHref(s.phone);
  const first = s.links?.[0];
  return first?.url || "";
}

function _normalizeServiceTarget(label) {
  const t = String(label || "").trim();
  if (!t) return "";
  const lower = t.toLowerCase();
  if (lower === "whatsapp" || lower === "open whatsapp") return "WhatsApp";
  if (lower === "website" || lower === "open website") return "website";
  return t;
}

function _normalizeUrlForMatch(url) {
  try {
    const parsed = new URL(url, window.location.origin);
    return parsed.href.replace(/\/$/, "");
  } catch {
    return String(url || "").trim();
  }
}

function _resolveServiceClickUrl(service, explicitTarget, rawUrl) {
  let url = String(rawUrl || "").trim();
  if (url && url !== "#") return url;
  if (!service) return "";
  const target = _normalizeServiceTarget(explicitTarget).toLowerCase();
  if (target === "whatsapp" && service.phone) return whatsappHref(service.phone);
  if (target === "website" && service.url) return service.url;
  if (service.links?.length && target) {
    const byLabel = service.links.find(
      (l) => l.label && l.label.toLowerCase() === target,
    );
    if (byLabel?.url) return byLabel.url;
  }
  return primaryServiceContact(service);
}

function _resolveServiceClickTarget(explicitTarget, url, service) {
  const fromButton = _normalizeServiceTarget(explicitTarget);
  if (fromButton) return fromButton;
  const resolvedUrl = String(url || "").trim();
  if (service?.links?.length && resolvedUrl) {
    const needle = _normalizeUrlForMatch(resolvedUrl);
    const match = service.links.find(
      (l) => l.url && _normalizeUrlForMatch(l.url) === needle,
    );
    if (match?.label) return _normalizeServiceTarget(match.label) || match.label;
  }
  return inferServiceTargetFromUrl(resolvedUrl);
}

function inferServiceTargetFromUrl(url) {
  if (!url) return "";
  try {
    const parsed = new URL(url, window.location.origin);
    if (parsed.protocol === "whatsapp:") return "WhatsApp";
    const host = parsed.hostname.toLowerCase();
    if (host === "wa.me" || host === "api.whatsapp.com") return "WhatsApp";
    if (host === "t.me" || host === "telegram.me" || host.endsWith(".t.me")) return "Telegram";
    if (parsed.protocol === "http:" || parsed.protocol === "https:") return "website";
  } catch {
    /* ignore invalid URLs */
  }
  return "";
}

function _serviceLinksHtml(s) {
  if (!s.links || !s.links.length) return "";
  return s.links
    .map(
      (l) => `
      <a class="link-item service-link"
         data-service-id="${s.id}"
         data-service-context="${esc(s._ctx || "list")}"
         data-service-target="${esc(l.label)}"
         data-url="${esc(l.url)}"
         href="${_linkHref(l.url)}">
        <span class="link-item-main">
          <span class="link-badge badge-other">OT</span>
          <span class="link-label">${esc(l.label)}</span>
        </span>
      </a>`,
    )
    .join("");
}

function _serviceContactRows(s, context) {
  const rows = [];
  if (s.phone) {
    const whatsapp = whatsappHref(s.phone);
    rows.push(`
      <a class="link-item service-link"
         data-service-id="${s.id}"
         data-service-context="${esc(context)}"
         data-service-target="WhatsApp"
         data-url="${esc(whatsapp)}"
         href="${_linkHref(whatsapp)}">
        <span class="link-item-main">
          <span class="link-badge badge-other">💬</span>
          <span class="link-label">Open WhatsApp</span>
        </span>
      </a>`);
  }
  if (s.url) {
    rows.push(`
      <a class="link-item service-link"
         data-service-id="${s.id}"
         data-service-context="${esc(context)}"
         data-service-target="website"
         data-url="${esc(s.url)}"
         href="${_linkHref(s.url)}">
        <span class="link-item-main">
          <span class="link-badge badge-other">🔗</span>
          <span class="link-label">Open website</span>
        </span>
      </a>`);
  }
  return rows.join("");
}

function _serviceVisual(s) {
  if (s.logo_url) {
    return `<img class="service-logo" src="${esc(s.logo_url)}" alt="" loading="lazy" referrerpolicy="no-referrer" />`;
  }
  return `<span class="service-logo-fallback" aria-hidden="true">${esc(s.emoji || "🤝")}</span>`;
}

function buildServiceCard(s, context = "list") {
  const code = s.category || "Community";
  const title = s.title || "Community Service";
  const subtitle = s.owner_name ? esc(s.owner_name) : "Community service";
  const desc = (s.description || "").trim();
  const contactRows = _serviceContactRows(s, context);
  const linksHtml = _serviceLinksHtml({ ...s, _ctx: context });

  return `
    <div class="course-card service-card" data-service-id="${s.id}" data-service-context="${esc(context)}">
      <div class="course-header" data-toggle-service="${s.id}">
        <div class="service-header-main">
          ${_serviceVisual(s)}
          <div class="service-header-text">
            <h2 class="course-name">${esc(title)}</h2>
            <p class="service-owner">${subtitle}</p>
          </div>
        </div>
        <div class="course-header-side">
          <div class="course-header-tags">
            <span class="optional-tag">COMMUNITY</span>
            <h3 class="course-code">${esc(code)}</h3>
          </div>
          <span class="course-chev" aria-hidden="true">›</span>
        </div>
      </div>
      <div class="links-list">
        ${desc ? `<p class="service-desc">${esc(desc)}</p>` : ""}
        ${contactRows || `<div class="service-desc muted">No contact info yet</div>`}
        ${linksHtml}
      </div>
    </div>`;
}

function buildServicePickCard(s) {
  const emoji = s.emoji || "🤝";
  const blurb = (s.description || "").trim() || s.owner_name || s.category || "Community service";
  return `
    <button type="button" class="pick-card service-pick-card" data-service-id="${s.id}" data-service-context="home">
      <span>
        <span class="pick-title">${s.logo_url ? "" : emoji + " "}${esc(s.title)}</span>
        <small>${esc(blurb.length > 72 ? blurb.slice(0, 72) + "…" : blurb)}</small>
      </span>
      <span class="pick-chev">›</span>
    </button>`;
}

function buildServiceSemButton(s) {
  return `
    <button type="button" class="mobile-sem-btn service-sem-btn" data-service-id="${s.id}" data-service-context="semester">
      ${esc(s.emoji || "🤝")} ${esc(s.title)}
    </button>`;
}

function intersperse(cards, services, context = "list") {
  if (!services || !services.length || !cards || !cards.length) return cards || [];
  const out = [];
  const step = 4;
  let si = 0;
  for (let i = 0; i < cards.length; i++) {
    out.push(cards[i]);
    if ((i + 1) % step === 0 && si < services.length) {
      out.push(buildServiceCard(services[si++], context));
    }
  }
  if (si < services.length) {
    out.push(buildServiceCard(services[si], context));
  }
  return out;
}

function trackServiceClick(serviceId, context, target, url) {
  if (!serviceId || AppState.adminLoggedIn) return;
  const s = AppState.dbServices.find((x) => x.id === serviceId);
  const resolvedUrl = _resolveServiceClickUrl(s, target, url);
  const linkTarget = _resolveServiceClickTarget(target, resolvedUrl, s);
  apiRequest("/api/service_clicks", {
    method: "POST",
    body: {
      service_id: serviceId,
      page_context: context || "list",
      link_target: linkTarget,
      url: resolvedUrl,
      clicked_url: resolvedUrl,
    },
  }).catch((e) => {
    if (e?.status === 401) window.onStudentTokenRejected?.();
    else if (e?.message) console.error("[service click]", e.message);
  });
}

function confirmServiceLink(serviceId, rawUrl, context, target) {
  const s = AppState.dbServices.find((x) => x.id === serviceId);
  let url = _resolveServiceClickUrl(s, target, rawUrl);
  if (!url) return;
  let parsed;
  try {
    parsed = new URL(url, window.location.origin);
  } catch (err) {
    return;
  }
  const linkTarget = _resolveServiceClickTarget(target, parsed.href, s);
  const isWhatsApp = linkTarget === "WhatsApp" || _isWhatsAppUrl(parsed);
  if (!isWhatsApp && !["http:", "https:"].includes(parsed.protocol)) return;

  const heading = isWhatsApp ? "Message on WhatsApp" : "Open community service";
  const actionLabel = isWhatsApp ? "Open WhatsApp ↗" : "Open Website ↗";
  const display = isWhatsApp ? (s?.phone || parsed.pathname.replace(/^\//, "")) : url;

  const box = document.createElement("div");
  box.innerHTML = `<h2>${esc("🤝 Community Service")}</h2>
  <p style="color:var(--muted);margin-top:8px;font-size:1rem;">${esc(heading)}:</p>
  <p style="word-break:break-all;font-family:monospace;background:var(--bg3);padding:8px;border-radius:4px;margin:10px 0;font-size:.85rem;">${esc(display)}</p>
  <div class="modal-actions">
    <button class="btn btn-ghost" onclick="closeModal()">Cancel</button>
    <button class="btn" id="openServiceLinkBtn">${esc(actionLabel)}</button>
  </div>`;

  document.getElementById("modalBox").innerHTML = "";
  document.getElementById("modalBox").appendChild(box);
  document.getElementById("modal").classList.add("open");

  document.getElementById("openServiceLinkBtn").addEventListener("click", () => {
    trackServiceClick(serviceId, context, linkTarget, parsed.href);
    closeModal();
    window.open(parsed.href, "_blank", "noopener,noreferrer");
  });
}

function renderMobileCommunity() {
  const container = document.getElementById("coursesOutput");
  if (!container) return;
  setSectionHint(COMMUNITY_HINT_CARD);
  const services = shuffleServices(`community-mobile:${performance.now()}`);
  if (!services.length) {
    container.innerHTML = `
      <button type="button" class="mobile-back" data-mobile-back="program">← Programs</button>
      <div class="empty">No community services yet.</div>`;
    return;
  }
  const cards = services.map((s) => buildServiceCard(s, "community"));
  container.innerHTML = `
    <button type="button" class="mobile-back" data-mobile-back="program">← Programs</button>
    <div class="mobile-section-label">Community Services</div>
    <div class="courses-grid">${cards.join("")}</div>`;
}

function renderDesktopCommunity() {
  const output = document.getElementById("coursesOutput");
  const extra = document.getElementById("extraSection");
  const sidebar = document.getElementById("serviceSidebar");
  if (!output || !extra) return;
  document.querySelector(".filter-row").style.display = "none";
  output.style.display = "";
  extra.style.display = "none";
  if (sidebar) sidebar.style.display = "none";
  const services = shuffleServices(`community:${performance.now()}`);
  setSectionHint("");
  if (!services.length) {
    output.innerHTML = `
      <div class="home-section">
        ${homeSectionHeading("🤝 Community Services")}
        ${sectionInlineHintHtml(COMMUNITY_PROMOTE_HINT)}
        <div class="empty">No community services yet.</div>
      </div>`;
    return;
  }
  const cards = services.map((s) => buildServiceCard(s, "community"));
  output.innerHTML = `
    <div class="home-section">
      ${homeSectionHeading("🤝 Community Services")}
      ${sectionInlineHintHtml(COMMUNITY_PROMOTE_HINT)}
      <div class="courses-grid">${cards.join("")}</div>
    </div>`;
}

function renderDesktopServiceSidebar() {
  const sidebar = document.getElementById("serviceSidebar");
  if (!sidebar) return;
  const services = pickServices(SIDEBAR_SERVICE_COUNT, "sidebar");
  if (!services.length) {
    sidebar.innerHTML = "";
    sidebar.style.display = "none";
    return;
  }
  sidebar.style.display = "";
  sidebar.innerHTML = `
    <div class="service-sidebar-title">🤝 Community</div>
    ${services
      .map(
        (s) => `
      <div class="service-sidebar-card" data-service-id="${s.id}" data-service-context="sidebar">
        <div class="service-sidebar-emoji">${s.logo_url
          ? `<img class="service-sidebar-logo" src="${esc(s.logo_url)}" alt="" loading="lazy" referrerpolicy="no-referrer" />`
          : esc(s.emoji || "🤝")}</div>
        <div class="service-sidebar-body">
          <div class="service-sidebar-name">${esc(s.title)}</div>
          <div class="service-sidebar-meta">${esc((s.description || "").trim() || s.category || "Community")}</div>
        </div>
      </div>`,
      )
      .join("")}
    <a href="#" data-view="community" class="service-sidebar-link">See all →</a>`;
}

function addCommunityTab() {
  const progTabs = document.getElementById("progTabs");
  if (!progTabs) return;
  const existing = progTabs.querySelector('[data-prog-tab="community"]');
  if (existing) return;
  const btn = document.createElement("button");
  btn.className = "prog-tab";
  btn.dataset.progTab = "community";
  btn.textContent = "🤝 Community";
  progTabs.appendChild(btn);
}

function highlightServiceCard(serviceId) {
  if (!serviceId) return;
  document.querySelectorAll(".service-card.service-highlight").forEach((el) => {
    el.classList.remove("service-highlight");
  });
  const card = document.querySelector(`.service-card[data-service-id="${serviceId}"]`);
  if (!card) return;
  card.classList.add("service-highlight");
  card.scrollIntoView({ behavior: "smooth", block: "center" });
  window.setTimeout(() => card.classList.remove("service-highlight"), 4500);
}

function selectCommunity(serviceId = null) {
  setHomeLegendVisible(true);
  if (window.showView) window.showView("home");
  if (isMobileView()) {
    renderMobileCommunity();
    if (serviceId) window.requestAnimationFrame(() => highlightServiceCard(serviceId));
    return;
  }
  AppState.currentProg = "community";
  AppState.currentYear = "all";
  AppState.currentSem = "all";
  renderProgTabs();
  renderDesktopCommunity();
  if (serviceId) {
    window.requestAnimationFrame(() => highlightServiceCard(serviceId));
  }
}

window.inferServiceTargetFromUrl = inferServiceTargetFromUrl;
window.trackServiceClick = trackServiceClick;
window.confirmServiceLink = confirmServiceLink;
window.primaryServiceContact = primaryServiceContact;
window.renderMobileCommunity = renderMobileCommunity;
window.renderDesktopCommunity = renderDesktopCommunity;
window.renderDesktopServiceSidebar = renderDesktopServiceSidebar;
window.addCommunityTab = addCommunityTab;
window.selectCommunity = selectCommunity;
window.highlightServiceCard = highlightServiceCard;
window.pickServices = pickServices;
window.pickRotatingService = pickRotatingService;
window.shuffleServices = shuffleServices;
window.buildServiceCard = buildServiceCard;
window.buildServicePickCard = buildServicePickCard;
window.buildServiceSemButton = buildServiceSemButton;
window.intersperse = intersperse;

export {
  pickServices,
  pickRotatingService,
  shuffleServices,
  buildServiceCard,
  buildServicePickCard,
  buildServiceSemButton,
  intersperse,
  inferServiceTargetFromUrl,
  trackServiceClick,
  confirmServiceLink,
  primaryServiceContact,
  whatsappHref,
  renderMobileCommunity,
  renderDesktopCommunity,
  renderDesktopServiceSidebar,
  addCommunityTab,
  selectCommunity,
  highlightServiceCard,
};
