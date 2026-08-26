import { AppState } from "./state.js";
import { apiRequest } from "./supabase.js";

/** @type {AbortController | null} */
let _controller = null;

function modelContext() {
  // Scanners and Chrome EPP expose navigator.modelContext; newer drafts use document.
  if (typeof navigator !== "undefined" && navigator.modelContext) {
    return navigator.modelContext;
  }
  if (typeof document !== "undefined" && document.modelContext) {
    return document.modelContext;
  }
  return null;
}

async function ensureContent() {
  if (AppState.dbPrograms?.length) return;
  if (typeof window.loadAll === "function") {
    await window.loadAll();
    return;
  }
  const data = await apiRequest("/api/content");
  // loadAll is preferred; without it, tools still answer from a one-shot fetch.
  AppState._webmcpContent = data;
}

function allCourses() {
  const out = [];
  if (AppState.dbPrograms?.length) {
    for (const p of AppState.dbPrograms) {
      for (const y of p.years || []) {
        for (const s of y.sems || []) {
          for (const c of s.courses || []) {
            out.push({
              id: c.id,
              name: c.name,
              code: c.code,
              program: p.name,
              year: y.name,
              semester: s.name,
              links: (c.links || []).map((l) => ({
                id: l.id,
                label: l.label,
                url: l.url,
                type: l.type,
              })),
            });
          }
        }
      }
    }
    return out;
  }
  const raw = AppState._webmcpContent;
  if (!raw?.courses) return out;
  const progById = new Map((raw.programs || []).map((p) => [p.id, p]));
  const yearById = new Map((raw.years || []).map((y) => [y.id, y]));
  const semById = new Map((raw.semesters || []).map((s) => [s.id, s]));
  const linksByCourse = new Map();
  for (const l of raw.links || []) {
    if (!linksByCourse.has(l.course_id)) linksByCourse.set(l.course_id, []);
    linksByCourse.get(l.course_id).push({
      id: l.id,
      label: l.label,
      url: l.url,
      type: l.type,
    });
  }
  for (const c of raw.courses) {
    const sem = semById.get(c.semester_id);
    const year = sem ? yearById.get(sem.year_id) : null;
    const prog = year ? progById.get(year.program_id) : null;
    out.push({
      id: c.id,
      name: c.name,
      code: c.code,
      program: prog?.name || null,
      year: year?.name || null,
      semester: sem?.name || null,
      links: linksByCourse.get(c.id) || [],
    });
  }
  return out;
}

function matchQuery(course, q) {
  if (!q) return true;
  const needle = q.toLowerCase();
  return (
    (course.name || "").toLowerCase().includes(needle) ||
    (course.code || "").toLowerCase().includes(needle)
  );
}

function toolDefs() {
  return [
    {
      name: "search_courses",
      description:
        "Search Info Links courses by name or code (e.g. NFA035, networks). Updates the on-page search and returns matching courses with program path.",
      inputSchema: {
        type: "object",
        properties: {
          query: {
            type: "string",
            description: "Course name or code fragment to search for.",
          },
          limit: {
            type: "number",
            description: "Max results to return (default 20).",
          },
        },
        required: ["query"],
      },
      annotations: { readOnlyHint: true },
      async execute(input = {}) {
        await ensureContent();
        const query = String(input.query || "").trim();
        const limit = Math.min(Math.max(Number(input.limit) || 20, 1), 50);
        const searchInput = document.getElementById("searchInput");
        if (searchInput) {
          searchInput.value = query;
          window.onSearch?.();
        }
        if (typeof window.showView === "function") window.showView("home");
        const results = allCourses()
          .filter((c) => matchQuery(c, query))
          .slice(0, limit)
          .map(({ links, ...rest }) => ({
            ...rest,
            link_count: links.length,
          }));
        return { query, count: results.length, results };
      },
    },
    {
      name: "list_programs",
      description:
        "List academic programs available on Info Links (CNAM Lebanon), plus Extra resources and Favorites tabs.",
      inputSchema: { type: "object", properties: {} },
      annotations: { readOnlyHint: true },
      async execute() {
        await ensureContent();
        const programs = (AppState.dbPrograms || []).map((p) => ({
          id: p.id,
          name: p.name,
          slug: p.slug,
        }));
        if (!programs.length && AppState._webmcpContent?.programs) {
          return {
            programs: AppState._webmcpContent.programs.map((p) => ({
              id: p.id,
              name: p.name,
              slug: p.slug,
            })),
          };
        }
        return {
          programs,
          tabs: ["all", "extra", "favorites"],
        };
      },
    },
    {
      name: "get_course",
      description:
        "Retrieve one course by id, exact code, or name, including material links (Drive, Classroom, Telegram, etc.). Prefer search_courses when unsure of the exact title.",
      inputSchema: {
        type: "object",
        properties: {
          code: { type: "string", description: "Course code, e.g. NFA035." },
          name: { type: "string", description: "Course name (partial OK)." },
          id: { type: "number", description: "Internal course id." },
        },
      },
      annotations: { readOnlyHint: true },
      async execute(input = {}) {
        await ensureContent();
        const courses = allCourses();
        let hit = null;
        if (input.id != null) {
          hit = courses.find((c) => c.id === Number(input.id));
        }
        if (!hit && input.code) {
          const code = String(input.code).toLowerCase().trim();
          hit = courses.find((c) => (c.code || "").toLowerCase() === code);
        }
        if (!hit && input.name) {
          const name = String(input.name).toLowerCase().trim();
          hit =
            courses.find((c) => (c.name || "").toLowerCase() === name) ||
            courses.find((c) => (c.name || "").toLowerCase().includes(name));
        }
        if (!hit) {
          return { found: false, message: "No matching course." };
        }
        return { found: true, course: hit };
      },
    },
    {
      name: "navigate",
      description:
        "Navigate the Info Links SPA to a main view: home (course browser), about, or admin login screen.",
      inputSchema: {
        type: "object",
        properties: {
          view: {
            type: "string",
            enum: ["home", "about", "admin"],
            description: "Target view id.",
          },
          program: {
            type: "string",
            description:
              "Optional program tab: all | extra | favorites | numeric program id.",
          },
        },
        required: ["view"],
      },
      annotations: { readOnlyHint: false },
      async execute(input = {}) {
        const view = String(input.view || "home");
        if (typeof window.showView === "function") {
          window.showView(view);
        } else {
          return { ok: false, message: "Navigation not ready yet." };
        }
        if (view === "home" && input.program != null && window.selectProg) {
          const raw = String(input.program);
          const prog =
            raw === "all" || raw === "extra" || raw === "favorites"
              ? raw
              : Number(raw);
          window.selectProg(prog);
        }
        return { ok: true, view, program: input.program ?? null };
      },
    },
  ];
}

/**
 * Register WebMCP tools for in-browser agents (WebMCP / modelContext API).
 * Uses AbortController so tools unregister when the page is torn down.
 */
async function initWebMCP() {
  const ctx = modelContext();
  if (!ctx || typeof ctx.registerTool !== "function") {
    return null;
  }

  if (_controller) {
    _controller.abort();
  }
  _controller = new AbortController();
  const { signal } = _controller;

  for (const tool of toolDefs()) {
    // Prefer navigator.modelContext.registerTool when present (isitagentready / Chrome).
    if (navigator.modelContext?.registerTool) {
      await navigator.modelContext.registerTool(tool, { signal });
    } else {
      await ctx.registerTool(tool, { signal });
    }
  }

  const abort = () => {
    if (_controller) {
      _controller.abort();
      _controller = null;
    }
  };
  window.addEventListener("pagehide", abort, { once: true });
  return _controller;
}

export { initWebMCP };
