/**
 * Adaptive PWA “Install app” UI.
 * Does not register or modify service workers / manifests — install prompt only.
 */

import { openModal, closeModal } from "./modals.js";

const SITE_NAME = "Info Links";
const DISMISS_KEY = "infolinks_pwa_install_dismissed";
// Chromium on Android fires beforeinstallprompt only after the service worker
// is active and engagement checks pass — that can take several seconds (or a
// second visit). Keep listening indefinitely; this timeout only relabels the
// still-pending button as a menu-install fallback, it never stops listening.
const UNSUPPORTED_WAIT_MS = 6000;

let deferredPrompt = null;
let buttonVisibleLogged = false;
let listenersAttached = false;
let mode = "hidden"; // "pending" | "native" | "ios-safari" | "ios-other" | "unsupported" | "hidden"

function trackPwaEvent(name, detail = {}) {
  console.log(`[pwa] ${name}`, detail);
}

function recordPwaInstall() {
  trackPwaEvent("pwa_installed");
  // trackPwaInstall dedupes in-flight/per-user, so duplicate appinstalled
  // signals (native + early stub) still record only once.
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

function isAndroidDevice() {
  return /android/i.test(navigator.userAgent || "");
}

function updateBannerCopy(banner) {
  const hint = banner.querySelector(".pwa-install-banner-hint");
  if (!hint) return;
  if (mode === "native" || mode === "pending") {
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
  if (isAndroidDevice()) {
    return `<h2>Install ${SITE_NAME}</h2>
    <p class="pwa-install-lead">This browser installs apps from its menu:</p>
    <ol class="pwa-install-steps">
      <li>Tap the <strong>⋮ menu</strong> button in your browser.</li>
      <li>Tap <strong>Install app</strong> or <strong>Add to Home screen</strong>.</li>
      <li>Tap <strong>Install</strong> / <strong>Add</strong> to confirm.</li>
    </ol>
    <div class="modal-actions">
      <button type="button" class="btn btn-primary" data-action="pwaInstallClose">Got it</button>
    </div>`;
  }
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
  // A stored prompt wins in any mode: it may have arrived after the fallback
  // label was shown (late beforeinstallprompt on Android Chromium).
  if (deferredPrompt) {
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
  // "pending" (prompt may still arrive) and "unsupported" (e.g. Firefox,
  // Vivaldi, Quetta — no beforeinstallprompt at all) both use the menu steps.
  openModal(unsupportedInstructionsHtml());
}

function onBeforeInstallPrompt(e) {
  e.preventDefault();
  deferredPrompt = e;
  showInstallUi("native");
}

function adoptEarlyPrompt() {
  const early = window.__pwaInstallState?.prompt;
  if (early && !deferredPrompt) {
    deferredPrompt = early;
    showInstallUi("native");
    return true;
  }
  return !!deferredPrompt;
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
  // No beforeinstallprompt yet (e.g. Firefox, Vivaldi, Quetta) — menu steps.
  // Keep listening: a late prompt still upgrades this via onBeforeInstallPrompt.
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

  if (!listenersAttached) {
    listenersAttached = true;
    window.addEventListener("beforeinstallprompt", onBeforeInstallPrompt);
    window.addEventListener("appinstalled", onAppInstalled);
    // Events captured by the <head> stub before this bundle loaded.
    window.addEventListener("pwa:prompt-ready", adoptEarlyPrompt);
    window.addEventListener("pwa:appinstalled", onAppInstalled);
  }

  // Install happened before listeners/session were ready (early stub flag).
  if (window.__pwaInstallState?.installed) {
    recordPwaInstall();
    hideInstallUi();
    return;
  }

  // Chromium fires appinstalled; iOS/Home Screen is detected on first standalone open.
  if (isStandalone()) {
    recordPwaInstall();
    hideInstallUi();
    return;
  }

  // Prompt fired before this module loaded — use the stored event.
  if (adoptEarlyPrompt()) return;

  // iOS never fires beforeinstallprompt — show instructions UI right away
  if (isIosDevice()) {
    showInstallUi(isIosSafari() ? "ios-safari" : "ios-other");
    return;
  }

  // Show the button immediately as "pending": if beforeinstallprompt arrives
  // while the user browses (common on Android), clicks switch to the native
  // dialog automatically. Only after the wait do we label it menu-fallback.
  if (mode === "hidden") {
    showInstallUi("pending");
    window.setTimeout(resolveFallbackMode, UNSUPPORTED_WAIT_MS);
  }
}

window.pwaInstallClick = handleInstallClick;
window.pwaInstallDismiss = dismissBanner;
window.pwaInstallClose = () => closeModal();

export { initPwaInstall, handleInstallClick, dismissBanner };
