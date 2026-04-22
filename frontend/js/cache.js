// ===================== CACHE =====================
const _CACHE_KEY = "infolinks_data";
const _CACHE_TS_KEY = "infolinks_cache_ts";
const _CACHE_TTL = 60 * 60 * 1000; // 1 hour

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
    const ts = localStorage.getItem(_CACHE_TS_KEY);
    if (!ts || Date.now() - parseInt(ts) > _CACHE_TTL) return null;
    const raw = localStorage.getItem(_CACHE_KEY);
    return raw ? JSON.parse(raw) : null;
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