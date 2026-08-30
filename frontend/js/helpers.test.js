import { describe, it, expect, beforeEach } from "vitest";
import { _saveCache, _loadCache, _clearCache } from "./cache.js";
import { esc } from "./ui.js";
import {
  whatsappHref,
  primaryServiceContact,
  inferServiceTargetFromUrl,
  intersperse,
} from "./services.js";

describe("cache", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("saves and loads payload", () => {
    _saveCache({ programs: [{ id: 1 }] });
    const loaded = _loadCache();
    expect(loaded).not.toBeNull();
    expect(loaded.data.programs[0].id).toBe(1);
    expect(loaded.stale).toBe(false);
  });

  it("returns null when empty", () => {
    expect(_loadCache()).toBeNull();
  });

  it("clears stored entries", () => {
    _saveCache({ ok: true });
    _clearCache();
    expect(_loadCache()).toBeNull();
  });
});

describe("esc", () => {
  it("escapes HTML special characters", () => {
    expect(esc(`<img src="x" onerror='alert(1)'>`)).toBe(
      "&lt;img src=&quot;x&quot; onerror=&#039;alert(1)&#039;&gt;",
    );
  });

  it("returns empty string for falsy input", () => {
    expect(esc("")).toBe("");
    expect(esc(null)).toBe("");
  });
});

describe("whatsappHref", () => {
  it("normalizes phone digits to wa.me", () => {
    expect(whatsappHref("+961 71 123 456")).toBe("https://wa.me/96171123456");
  });

  it("passes through existing wa.me links", () => {
    expect(whatsappHref("https://wa.me/96171123456")).toBe(
      "https://wa.me/96171123456",
    );
  });

  it("returns empty for blank input", () => {
    expect(whatsappHref("")).toBe("");
    expect(whatsappHref("abc")).toBe("");
  });
});

describe("primaryServiceContact", () => {
  it("prefers url, then phone, then first link", () => {
    expect(primaryServiceContact({ url: "https://t.me/x" })).toBe(
      "https://t.me/x",
    );
    expect(primaryServiceContact({ phone: "96171123456" })).toBe(
      "https://wa.me/96171123456",
    );
    expect(
      primaryServiceContact({ links: [{ label: "Site", url: "https://a.test" }] }),
    ).toBe("https://a.test");
  });
});

describe("inferServiceTargetFromUrl", () => {
  it("classifies common hosts", () => {
    expect(inferServiceTargetFromUrl("https://t.me/cnam")).toBe("Telegram");
    expect(inferServiceTargetFromUrl("https://wa.me/9617")).toBe("WhatsApp");
    expect(inferServiceTargetFromUrl("https://example.com")).toBe("website");
    expect(inferServiceTargetFromUrl("")).toBe("");
  });
});

describe("intersperse", () => {
  it("inserts a service card after every fourth course card", () => {
    const cards = ["c1", "c2", "c3", "c4"];
    const services = [{ id: 1, title: "Tutoring", emoji: "📚", links: [] }];
    const out = intersperse(cards, services, "list");
    expect(out.slice(0, 4)).toEqual(cards);
    expect(out[4]).toContain("Tutoring");
    expect(out).toHaveLength(5);
  });

  it("returns cards unchanged when no services", () => {
    expect(intersperse(["a", "b"], [], "list")).toEqual(["a", "b"]);
  });
});
