import { AppState } from "./state.js";
import { esc } from "./ui.js";

const STEP_ORDER = ["prog", "year", "sem", "course"];

const PLACEHOLDERS = {
  prog: "Select program…",
  year: "Select year…",
  sem: "Select semester…",
  course: "Select course…",
};

function offerings() {
  return AppState.dbPrograms || [];
}

function stepEl(prefix, level) {
  return document.getElementById(`${prefix}Step-${level}`);
}

function selectEl(prefix, level) {
  const ids = {
    prog: `${prefix}Prog`,
    year: `${prefix}Year`,
    sem: `${prefix}Sem`,
    course: `${prefix}CoursePick`,
  };
  return document.getElementById(ids[level]);
}

function setStepVisible(prefix, level, visible) {
  const wrap = stepEl(prefix, level);
  if (!wrap) return;
  wrap.hidden = !visible;
  if (!visible) {
    const sel = selectEl(prefix, level);
    if (sel) {
      fillSelect(sel, [], { placeholder: PLACEHOLDERS[level], disabled: true });
    }
  }
}

function hideStepsAfter(prefix, level) {
  const start = STEP_ORDER.indexOf(level);
  for (let i = start + 1; i < STEP_ORDER.length; i++) {
    setStepVisible(prefix, STEP_ORDER[i], false);
  }
}

function showStep(prefix, level, options, placeholder) {
  const wrap = stepEl(prefix, level);
  const sel = selectEl(prefix, level);
  if (!wrap || !sel) return;
  wrap.hidden = false;
  fillSelect(sel, options, {
    placeholder: placeholder || PLACEHOLDERS[level],
    disabled: false,
  });
  queueMicrotask(() => {
    try {
      sel.focus({ preventScroll: false });
    } catch {
      sel.focus();
    }
  });
}

function fillSelect(el, options, { placeholder, disabled = false, selected = "" } = {}) {
  if (!el) return;
  const opts = [];
  if (placeholder != null) {
    opts.push(`<option value="">${esc(placeholder)}</option>`);
  }
  options.forEach(([value, label]) => {
    const sel = String(selected) === String(value) ? " selected" : "";
    opts.push(`<option value="${esc(String(value))}"${sel}>${esc(label)}</option>`);
  });
  el.innerHTML = opts.join("");
  el.disabled = disabled || (options.length === 0 && !selected);
  if (!el.disabled && selected !== "" && selected != null) {
    el.value = String(selected);
  }
}

function notifyCourseChange(prefix) {
  if (typeof window.onHierarchyCourseChange === "function") {
    window.onHierarchyCourseChange(prefix);
  }
}

/** Populate cascading selects for a form prefix (r / c). */
function hierarchyCascade(prefix, changed) {
  const progEl = selectEl(prefix, "prog");
  const yearEl = selectEl(prefix, "year");
  const semEl = selectEl(prefix, "sem");
  const courseEl = selectEl(prefix, "course");

  if (changed === "init") {
    if (progEl) {
      fillSelect(
        progEl,
        offerings().map((p) => [p.id, p.name]),
        { placeholder: PLACEHOLDERS.prog, disabled: false },
      );
    }
    hideStepsAfter(prefix, "prog");
    const progStep = stepEl(prefix, "prog");
    if (progStep) progStep.hidden = false;
    notifyCourseChange(prefix);
    return;
  }

  if (changed === "prog") {
    hideStepsAfter(prefix, "prog");
    const progId = parseInt(progEl?.value, 10);
    if (!progId) {
      notifyCourseChange(prefix);
      return;
    }
    const prog = offerings().find((p) => p.id === progId);
    showStep(
      prefix,
      "year",
      (prog?.years || []).map((y) => [y.id, y.name]),
      PLACEHOLDERS.year,
    );
    notifyCourseChange(prefix);
    return;
  }

  if (changed === "year") {
    hideStepsAfter(prefix, "year");
    const progId = parseInt(progEl?.value, 10);
    const yearId = parseInt(yearEl?.value, 10);
    const prog = offerings().find((p) => p.id === progId);
    const year = prog?.years?.find((y) => y.id === yearId);
    if (!year) {
      notifyCourseChange(prefix);
      return;
    }
    showStep(
      prefix,
      "sem",
      (year.sems || []).map((s) => [s.id, s.name]),
      PLACEHOLDERS.sem,
    );
    notifyCourseChange(prefix);
    return;
  }

  if (changed === "sem") {
    hideStepsAfter(prefix, "sem");
    if (courseEl) {
      const progId = parseInt(progEl?.value, 10);
      const yearId = parseInt(yearEl?.value, 10);
      const semId = parseInt(semEl?.value, 10);
      const prog = offerings().find((p) => p.id === progId);
      const year = prog?.years?.find((y) => y.id === yearId);
      const sem = year?.sems?.find((s) => s.id === semId);
      if (!sem) {
        notifyCourseChange(prefix);
        return;
      }
      const courses = (sem.courses || []).map((c) => [
        c.id,
        c.code ? `${c.code} — ${c.name}` : c.name,
      ]);
      showStep(prefix, "course", courses, PLACEHOLDERS.course);
    }
    notifyCourseChange(prefix);
    return;
  }

  if (changed === "course") {
    notifyCourseChange(prefix);
  }
}

function hierarchyStepHtml(prefix, level, label, icon, selectId, onchange, { hidden = false } = {}) {
  return `<div class="hierarchy-step" id="${prefix}Step-${level}" ${hidden ? "hidden" : ""}>
    <label class="feedback-field-label" for="${selectId}"><span aria-hidden="true">${icon}</span> <span>${esc(label)}</span></label>
    <select id="${selectId}" ${hidden ? "disabled" : ""} onchange="${onchange}">
      <option value="">${esc(PLACEHOLDERS[level])}</option>
    </select>
  </div>`;
}

function hierarchySelectHtml(prefix, { includeCourse = false } = {}) {
  const courseBlock = includeCourse
    ? hierarchyStepHtml(
        prefix,
        "course",
        "Course",
        "📚",
        `${prefix}CoursePick`,
        `hierarchyCascade('${prefix}','course')`,
        { hidden: true },
      )
    : "";
  return `<div class="hierarchy-fields" id="${prefix}Hierarchy">
    ${hierarchyStepHtml(prefix, "prog", "Program", "🎓", `${prefix}Prog`, `hierarchyCascade('${prefix}','prog')`)}
    ${hierarchyStepHtml(prefix, "year", "Year", "📅", `${prefix}Year`, `hierarchyCascade('${prefix}','year')`, { hidden: true })}
    ${hierarchyStepHtml(prefix, "sem", "Semester", "🗓", `${prefix}Sem`, `hierarchyCascade('${prefix}','sem')`, { hidden: true })}
    ${courseBlock}
  </div>`;
}

function initHierarchyPicker(prefix, { includeCourse = false, selected = null } = {}) {
  hierarchyCascade(prefix, "init");
  if (!selected) return;

  const { programId, yearId, semesterId, courseId } = selected;
  const progEl = selectEl(prefix, "prog");
  if (programId && progEl) {
    progEl.value = String(programId);
    hierarchyCascade(prefix, "prog");
  }
  const yearEl = selectEl(prefix, "year");
  if (yearId && yearEl) {
    yearEl.value = String(yearId);
    hierarchyCascade(prefix, "year");
  }
  const semEl = selectEl(prefix, "sem");
  if (semesterId && semEl) {
    semEl.value = String(semesterId);
    hierarchyCascade(prefix, "sem");
  }
  if (includeCourse && courseId) {
    const courseEl = selectEl(prefix, "course");
    if (courseEl) {
      courseEl.value = String(courseId);
      hierarchyCascade(prefix, "course");
    }
  }
}

function selectedCourseId(prefix) {
  const v = selectEl(prefix, "course")?.value;
  return v ? parseInt(v, 10) : null;
}

function selectedCourse(prefix) {
  const id = selectedCourseId(prefix);
  if (!id) return null;
  return AppState.courseById?.get(id) || null;
}

function findCoursePath(courseId) {
  for (const prog of offerings()) {
    for (const year of prog.years || []) {
      for (const sem of year.sems || []) {
        if ((sem.courses || []).some((c) => c.id === courseId)) {
          return {
            programId: prog.id,
            yearId: year.id,
            semesterId: sem.id,
            courseId,
          };
        }
      }
    }
  }
  return null;
}

window.hierarchyCascade = hierarchyCascade;
window.initHierarchyPicker = initHierarchyPicker;

export {
  hierarchySelectHtml,
  hierarchyCascade,
  initHierarchyPicker,
  selectedCourseId,
  selectedCourse,
  findCoursePath,
};
