// ===================== ADMIN STUDENT DIRECTORY =====================
// Shared id → handle lookup so reports / contributions / feedback can show a
// Sender column without one request per row.
import { sb, logApiError } from "./supabase.js";
import { esc } from "./ui.js";

const DIRECTORY_PAGE_SIZE = 200;
const DIRECTORY_MAX_PAGES = 10;

const studentDirectory = new Map();
let _directoryLoaded = false;
let _directoryInFlight = null;

function studentHandleOf(user) {
  if (!user) return "";
  return (
    user.handle || `${user.first_name}_${user.last_name}_${user.number}`
  );
}

function rememberStudents(rows) {
  if (!Array.isArray(rows)) return;
  rows.forEach((u) => {
    if (u && u.id != null) studentDirectory.set(Number(u.id), studentHandleOf(u));
  });
}

/** Load (once) every student handle so sender names resolve locally. */
function loadStudentDirectory({ force = false } = {}) {
  if (_directoryLoaded && !force) return Promise.resolve(studentDirectory);
  if (_directoryInFlight) return _directoryInFlight;

  _directoryInFlight = (async () => {
    try {
      for (let page = 0; page < DIRECTORY_MAX_PAGES; page++) {
        const offset = page * DIRECTORY_PAGE_SIZE;
        const rows = await sb(
          `users?limit=${DIRECTORY_PAGE_SIZE}&offset=${offset}`,
          "GET",
        );
        if (!Array.isArray(rows) || !rows.length) break;
        rememberStudents(rows);
        if (rows.length < DIRECTORY_PAGE_SIZE) break;
      }
      _directoryLoaded = true;
    } catch (err) {
      logApiError(err, "adminStudentDirectory");
    } finally {
      _directoryInFlight = null;
    }
    return studentDirectory;
  })();

  return _directoryInFlight;
}

/** Table cell for the student behind a submission; legacy rows show a dash. */
function senderCell(userId) {
  if (userId === null || userId === undefined || userId === "") return "—";
  const id = Number(userId);
  if (!Number.isFinite(id)) return "—";
  const handle = studentDirectory.get(id);
  if (!handle) {
    return `<span style="color:var(--muted);">#${esc(String(userId))}</span>`;
  }
  return `<button class="action-btn" onclick="openAdminStudent(${id})" title="Open student history">${esc(handle)}</button>`;
}

function fmtDateTime(ts) {
  if (!ts) return "—";
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("en", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export {
  studentDirectory,
  studentHandleOf,
  rememberStudents,
  loadStudentDirectory,
  senderCell,
  fmtDateTime,
};
