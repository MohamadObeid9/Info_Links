/**
 * Adaptive PWA “Install app” UI.
 * Does not register or modify service workers / manifests — install prompt only.
 */

import { openModal, closeModal } from "./modals.js";

const SITE_NAME = "Info Links";
const DISMISS_KEY = "infolinks_pwa_install_dismissed";
const UNSUPPORTED_WAIT_MS = 1500;

let deferredPrompt = null;
let buttonVisibleLogged = false;
let mode = "hidden"; // "native" | "ios-safari" | "ios-other" | "unsupported" | "hidden"

function trackPwaEvent(name, detail = {}) {
  console.log(`[pwa] ${name}`, detail);
}

function recordPwaInstall() {
  trackPwaEvent("pwa_installed");
  window.trackPwaInstall?.();
}

function isStandalone() {
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    window.navigator.standalone === true
  );
}

function isIosDevice() {
  const ua = navigator.userAgent || "";
  if (/iphone|ipad|ipod/i.test(ua)) return true;
  // iPadOS 13+ reports as Mac with touch
  return navigator.platform === "MacIntel" && navigator.maxTouchPoints > 1;
}

function isIosSafari() {
  if (!isIosDevice()) return false;
  const ua = navigator.userAgent || "";
  // Chrome/Firefox/Edge on iOS use CriOS / FxiOS / EdgiOS
  if (/crios|fxios|edgios/i.test(ua)) return false;
  return /safari/i.test(ua);
}

function isBannerDismissed() {
  try {
    return localStorage.getItem(DISMISS_KEY) === "1";
  } catch {
    return false;
  }
}

function dismissBanner() {
  try {
    localStorage.setItem(DISMISS_KEY, "1");
  } catch {
    /* ignore */
  }
  const banner = document.getElementById("pwaInstallBanner");
  if (banner) banner.hidden = true;
}

function hideInstallUi() {
  mode = "hidden";
  deferredPrompt = null;
  const btn = document.getElementById("installAppBtn");
  const banner = document.getElementById("pwaInstallBanner");
  if (btn) btn.hidden = true;
  if (banner) banner.hidden = true;
}

function showInstallUi(nextMode) {
  if (isStandalone()) {
    hideInstallUi();
    return;
  }
  mode = nextMode;
  const btn = document.getElementById("installAppBtn");
  if (btn) btn.hidden = false;

  const banner = document.getElementById("pwaInstallBanner");
  if (banner && !isBannerDismissed()) {
    banner.hidden = false;
    updateBannerCopy(banner);
  }

  if (!buttonVisibleLogged) {
    buttonVisibleLogged = true;
    trackPwaEvent("pwa_install_button_visible", { mode: nextMode });
  }
}

function updateBannerCopy(banner) {
  const hint = banner.querySelector(".pwa-install-banner-hint");
  if (!hint) return;
  if (mode === "native") {
    hint.textContent = `Use ${SITE_NAME} like a normal app, without opening the browser.`;
  } else if (mode === "ios-safari" || mode === "ios-other") {
    hint.textContent = `Add ${SITE_NAME} to your Home Screen for quick access.`;
  } else {
    hint.textContent = "Install from your browser menu for quick access.";
  }
}

function iosSafariInstructionsHtml() {
  return `<h2>Install ${SITE_NAME}</h2>
    <p class="pwa-install-lead">Use ${SITE_NAME} like a normal app, without opening the browser.</p>
    <ol class="pwa-install-steps">
      <li>Tap the <strong>Share</strong> button in Safari.</li>
      <li>Tap <strong>Add to Home Screen</strong>.</li>
      <li>Tap <strong>Add</strong>.</li>
    </ol>
    <div class="modal-actions">
      <button type="button" class="btn btn-primary" data-action="pwaInstallClose">Got it</button>
    </div>`;
}

function iosOtherBrowserInstructionsHtml() {
  return `<h2>Install ${SITE_NAME}</h2>
    <p class="pwa-install-lead">Installation on iPhone and iPad works from Safari.</p>
    <ol class="pwa-install-steps">
      <li>Open this website in <strong>Safari</strong>.</li>
      <li>Tap the <strong>Share</strong> button.</li>
      <li>Tap <strong>Add to Home Screen</strong>, then <strong>Add</strong>.</li>
    </ol>
    <div class="modal-actions">
      <button type="button" class="btn btn-primary" data-action="pwaInstallClose">Got it</button>
    </div>`;
}

function unsupportedInstructionsHtml() {
  return `<h2>Install ${SITE_NAME}</h2>
    <p class="pwa-install-lead">Open your browser menu and choose “Install” or “Add to Home Screen”.</p>
    <div class="modal-actions">
      <button type="button" class="btn btn-primary" data-action="pwaInstallClose">Got it</button>
    </div>`;
}

async function promptNativeInstall() {
  if (!deferredPrompt) return;
  trackPwaEvent("pwa_install_clicked", { mode: "native" });
  const promptEvent = deferredPrompt;
  deferredPrompt = null;
  promptEvent.prompt();
  let outcome = "dismissed";
  try {
    const choice = await promptEvent.userChoice;
    outcome = choice?.outcome || "dismissed";
  } catch {
    outcome = "dismissed";
  }
  if (outcome === "accepted") {
    trackPwaEvent("pwa_install_accepted");
    hideInstallUi();
  } else {
    trackPwaEvent("pwa_install_dismissed");
    // Prompt was consumed; offer browser-menu instructions until BIP re-fires
    showInstallUi("unsupported");
  }
}

function handleInstallClick() {
  if (isStandalone()) {
    hideInstallUi();
    return;
  }
  if (mode === "native" && deferredPrompt) {
    promptNativeInstall();
    return;
  }
  trackPwaEvent("pwa_install_clicked", { mode });
  if (mode === "ios-safari") {
    openModal(iosSafariInstructionsHtml());
    return;
  }
  if (mode === "ios-other") {
    openModal(iosOtherBrowserInstructionsHtml());
    return;
  }
  if (mode === "unsupported") {
    openModal(unsupportedInstructionsHtml());
  }
}

function onBeforeInstallPrompt(e) {
  e.preventDefault();
  deferredPrompt = e;
  showInstallUi("native");
}

function onAppInstalled() {
  recordPwaInstall();
  hideInstallUi();
}

function resolveFallbackMode() {
  if (isStandalone() || deferredPrompt || mode === "native") return;
  if (isIosDevice()) {
    showInstallUi(isIosSafari() ? "ios-safari" : "ios-other");
    return;
  }
  // No beforeinstallprompt (e.g. Firefox) — show menu instructions, not a native prompt button
  showInstallUi("unsupported");
}

function ensureBannerMount() {
  if (document.getElementById("pwaInstallBanner")) return;
  const home = document.getElementById("view-home");
  if (!home) return;
  const banner = document.createElement("div");
  banner.id = "pwaInstallBanner";
  banner.className = "pwa-install-banner";
  banner.hidden = true;
  banner.innerHTML = `
    <button type="button" class="pwa-install-banner-cta" data-action="pwaInstall">Install app</button>
    <p class="pwa-install-banner-hint"></p>
    <button type="button" class="pwa-install-banner-dismiss" data-action="pwaInstallDismiss" aria-label="Dismiss install banner">✕</button>
  `;
  const hero = home.querySelector(".hero");
  if (hero && hero.nextSibling) {
    home.insertBefore(banner, hero.nextSibling);
  } else {
    home.prepend(banner);
  }
}

function ensureNavButton() {
  if (document.getElementById("installAppBtn")) return;
  const nav = document.querySelector("nav");
  if (!nav) return;

  // Keep the control outside the hamburger drawer so phones still see it
  // after the home banner is dismissed.
  let actions = document.getElementById("navActions");
  if (!actions) {
    actions = document.createElement("div");
    actions.id = "navActions";
    actions.className = "nav-actions";
    const hamburger = document.getElementById("hamburgerBtn");
    const navLinks = document.getElementById("navLinks");
    const anchor = hamburger || navLinks;
    if (anchor) nav.insertBefore(actions, anchor);
    else nav.appendChild(actions);

    // Desktop order: links → Install → (hidden hamburger)
    // Mobile header: Install → hamburger (links open as overlay)
    if (navLinks) actions.appendChild(navLinks);
    if (hamburger) actions.appendChild(hamburger);
  }

  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "install-app-btn";
  btn.id = "installAppBtn";
  btn.dataset.action = "pwaInstall";
  btn.hidden = true;
  btn.textContent = "Install app";

  const hamburger = document.getElementById("hamburgerBtn");
  if (hamburger && hamburger.parentElement === actions) {
    actions.insertBefore(btn, hamburger);
  } else {
    actions.appendChild(btn);
  }
}

function initPwaInstall() {
  ensureNavButton();
  ensureBannerMount();

  // Chromium fires appinstalled; iOS/Home Screen is detected on first standalone open.
  if (isStandalone()) {
    recordPwaInstall();
    hideInstallUi();
    return;
  }

  window.addEventListener("beforeinstallprompt", onBeforeInstallPrompt);
  window.addEventListener("appinstalled", onAppInstalled);

  // iOS never fires beforeinstallprompt — show instructions UI right away
  if (isIosDevice()) {
    showInstallUi(isIosSafari() ? "ios-safari" : "ios-other");
    return;
  }

  // Give Chromium a moment to fire beforeinstallprompt before showing menu fallback
  window.setTimeout(resolveFallbackMode, UNSUPPORTED_WAIT_MS);
}

window.pwaInstallClick = handleInstallClick;
window.pwaInstallDismiss = dismissBanner;
window.pwaInstallClose = () => closeModal();

export { initPwaInstall, handleInstallClick, dismissBanner };
