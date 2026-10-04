import { beforeAll, describe, expect, it } from "vitest";

beforeAll(() => {
  document.body.innerHTML = `
    <div id="modal"></div>
    <div id="view-admin">
      <div id="adminContent">
        <div class="admin-entity-card" id="course-card">
          <button type="button" class="admin-entity-toggle">
            <strong id="course-name">Algorithms</strong>
          </button>
          <div class="admin-link-list">course links</div>
        </div>
        <div class="admin-entity-card" id="extra-card">
          <button type="button" class="admin-entity-toggle">
            <span id="extra-name">Tips</span>
          </button>
          <div class="admin-link-list">extra links</div>
        </div>
      </div>
    </div>`;

  window.matchMedia = (query) => ({
    matches: String(query).includes("max-width"),
    media: String(query),
    addEventListener() {},
    removeEventListener() {},
    addListener() {},
    removeListener() {},
    dispatchEvent() {
      return false;
    },
  });
});

describe("admin mobile entity cards", () => {
  it("opens a course and an extra section from their header buttons", async () => {
    await import("./admin.js");

    const tap = (id) => {
      document.getElementById(id).dispatchEvent(new MouseEvent("click", { bubbles: true }));
    };

    tap("course-name");
    expect(document.getElementById("course-card").classList.contains("open")).toBe(true);

    tap("extra-name");
    expect(document.getElementById("extra-card").classList.contains("open")).toBe(true);
    expect(document.getElementById("course-card").classList.contains("open")).toBe(false);

    tap("extra-name");
    expect(document.getElementById("extra-card").classList.contains("open")).toBe(false);
  });
});
