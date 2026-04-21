// ===================== SKELETON SCREEN =====================

/**
 * Render a full skeleton in place of the home page content.
 * Called immediately on first load before any Supabase data arrives.
 *
 * Layout mirrors the real page:
 *   prog-tabs row → filter pills → (sem title + cards grid) × N
 */
function showSkeleton() {
  // ── Program tabs ──────────────────────────────────────────
  const tabWidths = [80, 120, 100, 140, 130]; // px, mimics real tab labels
  document.getElementById("progTabs").innerHTML = tabWidths
    .map((w) => `<div class="skel skel-tab" style="width:${w}px;"></div>`)
    .join("");

  // ── Year + Semester filter pills ──────────────────────────
  const filterWidths = [50, 80, 90, 70, 85, 75, 65];
  document.getElementById("yearFilters").innerHTML = filterWidths
    .slice(0, 4)
    .map(
      (w) => `<div class="skel skel-filter-pill" style="width:${w}px;"></div>`,
    )
    .join("");
  document.getElementById("semFilters").innerHTML = filterWidths
    .slice(3)
    .map(
      (w) => `<div class="skel skel-filter-pill" style="width:${w}px;"></div>`,
    )
    .join("");

  // ── Course cards ──────────────────────────────────────────
  // Show 2 semester blocks, each with a row of cards
  const blocks = [
    { cards: 3, linksPerCard: 2 },
    { cards: 4, linksPerCard: 3 },
    { cards: 3, linksPerCard: 2 },
  ];

  document.getElementById("coursesOutput").innerHTML = blocks
    .map(
      ({ cards, linksPerCard }) => `
            <div style="margin-bottom:32px;">
                <!-- Year heading -->
                <div class="skel" style="height:18px;width:140px;border-radius:4px;margin-bottom:16px;"></div>
                <!-- Semester label -->
                <div class="skel skel-sem-title"></div>
                <!-- Cards grid -->
                <div class="skel-grid">
                    ${Array.from(
                      { length: cards },
                      () => `
                        <div class="skel-card">
                            <div class="skel-card-header">
                                <div class="skel skel-card-title"></div>
                                <div class="skel skel-card-code"></div>
                            </div>
                            ${Array.from(
                              { length: linksPerCard },
                              () => `<div class="skel skel-link"></div>`,
                            ).join("")}
                        </div>
                    `,
                    ).join("")}
                </div>
            </div>
        `,
    )
    .join("");
}

/**
 * Remove all skeleton markup — called after real data has rendered.
 * (Currently a no-op because renderCourses/renderProgTabs overwrite
 *  the same containers, but kept here for clarity & future use.)
 */
function hideSkeleton() {
  // Real render functions replace innerHTML directly,
  // so skeletons are automatically replaced. Nothing extra needed.
}

function getAdminTableSkeleton() {
  return `
    <div style="margin-bottom: 20px;">
      <div class="skel" style="height: 38px; width: 100%; border-radius: 8px; margin-bottom: 16px;"></div>
      <table class="admin-table">
        <thead>
          <tr>
            <th><div class="skel" style="height: 16px; width: 60px; border-radius: 4px;"></div></th>
            <th><div class="skel" style="height: 16px; width: 80px; border-radius: 4px;"></div></th>
            <th><div class="skel" style="height: 16px; width: 100px; border-radius: 4px;"></div></th>
            <th><div class="skel" style="height: 16px; width: 50px; border-radius: 4px;"></div></th>
            <th><div class="skel" style="height: 16px; width: 60px; border-radius: 4px;"></div></th>
          </tr>
        </thead>
        <tbody>
          ${Array.from({length: 4}, () => `
            <tr>
              <td><div class="skel" style="height: 14px; width: 80px; border-radius: 4px;"></div></td>
              <td><div class="skel" style="height: 14px; width: 120px; border-radius: 4px;"></div></td>
              <td><div class="skel" style="height: 14px; width: 150px; border-radius: 4px;"></div></td>
              <td><div class="skel" style="height: 24px; width: 60px; border-radius: 12px;"></div></td>
              <td>
                <div class="skel" style="height: 28px; width: 70px; border-radius: 6px; display: inline-block;"></div>
                <div class="skel" style="height: 28px; width: 30px; border-radius: 6px; display: inline-block;"></div>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}

function getAdminAnalyticsSkeleton() {
  return `
    <div class="stat-grid">
      ${Array.from({length: 4}, () => `
        <div class="stat-card">
          <div class="skel" style="height: 32px; width: 60px; border-radius: 6px; margin-bottom: 8px;"></div>
          <div class="skel" style="height: 14px; width: 100px; border-radius: 4px;"></div>
        </div>
      `).join('')}
    </div>
    <div class="chart-wrap" style="margin-top: 20px;">
      <div class="skel" style="height: 18px; width: 150px; border-radius: 4px; margin-bottom: 24px;"></div>
      <div style="display: flex; gap: 8px; margin-bottom: 20px;">
        <div class="skel" style="height: 30px; width: 70px; border-radius: 20px;"></div>
        <div class="skel" style="height: 30px; width: 70px; border-radius: 20px;"></div>
        <div class="skel" style="height: 30px; width: 70px; border-radius: 20px;"></div>
      </div>
      <div style="height: 200px; width: 100%; border-radius: 8px;" class="skel"></div>
    </div>
  `;
}
window.showSkeleton = showSkeleton;
