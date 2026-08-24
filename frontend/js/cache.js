// ===================== CACHE =====================
const _CACHE_KEY = "infolinks_data";
const _CACHE_TS_KEY = "infolinks_cache_ts";
const _CACHE_TTL = 60 * 60 * 1000; // fresh window; stale entries still render while we revalidate

function _saveCache(data) {
  try {
    localStorage.setItem(_CACHE_KEY, JSON.stringify(data));
    localStorage.setItem(_CACHE_TS_KEY, Date.now().toString());
  } catch (e) {
    // localStorage full or unavailable — fail silently
  }
}

function _loadCache() {
  try {
    const raw = localStorage.getItem(_CACHE_KEY);
    if (!raw) return null;
    const data = JSON.parse(raw);
    if (!data || typeof data !== "object") return null;
    const ts = parseInt(localStorage.getItem(_CACHE_TS_KEY) || "0", 10);
    const stale = !ts || Date.now() - ts > _CACHE_TTL;
    return { data, stale };
  } catch (e) {
    return null;
  }
}

function _clearCache() {
  try {
    localStorage.removeItem(_CACHE_KEY);
    localStorage.removeItem(_CACHE_TS_KEY);
  } catch (e) {}
}

export { _saveCache, _loadCache, _clearCache };