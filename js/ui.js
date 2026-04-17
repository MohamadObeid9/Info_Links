// ===================== HELPERS =====================
function esc(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// ===================== THEME =====================
function toggleTheme() {
  isDark = !isDark;
  document.documentElement.setAttribute(
    "data-theme",
    isDark ? "dark" : "light",
  );
  document.getElementById("themeBtn").textContent = isDark ? "🌙" : "☀️";
}

// ===================== MOBILE =====================
function toggleMobileMenu() {
  document.getElementById("hamburgerBtn").classList.toggle("open");
  document.getElementById("navLinks").classList.toggle("mobile-open");
}
function toggleFilters() {
  document.getElementById("filterToggleBtn").classList.toggle("open");
  document.getElementById("filtersCollapsible").classList.toggle("open");
}
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("nav-btn")) {
    document.getElementById("hamburgerBtn").classList.remove("open");
    document.getElementById("navLinks").classList.remove("mobile-open");
  }
});

// Helper: find all courses that share the same code as the given courseId
function _findSharedCourses(courseId) {
  let targetCode = null;
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id === courseId) targetCode = c.code;
        }),
      ),
    ),
  );
  if (!targetCode) return { code: null, siblings: [] };
  const siblings = [];
  dbPrograms.forEach((p) =>
    p.years.forEach((y) =>
      y.sems.forEach((s) =>
        s.courses.forEach((c) => {
          if (c.id !== courseId && c.code === targetCode)
            siblings.push({
              ...c,
              prog: p.name,
              year: y.name,
              sem: s.name,
            });
        }),
      ),
    ),
  );
  return { code: targetCode, siblings };
}

// ===================== BADGE =====================
function getLinkBadge(type) {
  if (type === "telegram") return '<span class="link-badge badge-tg">TG</span>';
  if (type === "drive") return '<span class="link-badge badge-drive">GD</span>';
  if (type === "classroom")
    return '<span class="link-badge badge-classroom">GC</span>';
  if (type === "other") return '<span class="link-badge badge-other">OT</span>';
  return '<span class="link-badge badge-other">OT</span>';
}

// ===================== BANNER =====================
function toggleImportantNote() {
  document.getElementById("importantNote").classList.toggle("open");
}

const noteTranslations = {
  en: {
    title: "Important Note",
    content: `We are improving the site and need your help! We'll add a feature to list the content of each link. If
                you tried any link before and know its content, click on the <strong>Report / Contribute</strong>
                button. In the
                contribute section, list the name of the course and its code, then list the link number in the link
                section (e.g., link 1 / link 2 / link 3 ...). In the note
                section, add your info (e.g., whether it contains TD only, Cours only, videos only, sessions, or if
                there are any missing chapters). Do that for each link u know, and we'll appreciate your contributions.
                ❤️`
  },
  fr: {
    title: "Note Importante",
    content: `Nous améliorons le site et avons besoin de votre aide ! Nous allons ajouter une fonctionnalité pour lister le contenu de chaque lien. Si vous avez déjà essayé un lien et connaissez son contenu, cliquez sur le bouton <strong>Report / Contribute</strong>. Dans la section "contribute", indiquez le nom du cours et son code, puis le numéro du lien dans la section "link" (ex: link 1 / link 2 / link 3 ...). Dans la section "note", ajoutez vos informations (ex: TD uniquement, Cours uniquement, vidéos, ou s'il manque des chapitres). Faites cela pour chaque lien que vous connaissez, et nous apprécierons vos contributions. ❤️`
  },
  ar: {
    title: "ملاحظة هامة",
    content: `نحن نعمل على تحسين الموقع ونحتاج إلى مساعدتك! سنضيف ميزة لسرد محتوى كل رابط. إذا قمت بتجربة أي رابط من قبل وتعرف محتواه، فانقر فوق الزر <strong>Report / Contribute</strong>. في قسم المساهمة (contribute)، اذكر اسم المادة ورمزها، ثم اذكر رقم الرابط في قسم الرابط (على سبيل المثال، link 1 / link 2 / link 3 ...). في قسم الملاحظات (note)، أضف معلوماتك (على سبيل المثال، ما إذا كان يحتوي على TD فقط، أو Cours فقط، أو مقاطع فيديو فقط، أو إذا كانت هناك أي فصول مفقودة). افعل ذلك لكل رابط تعرفه، وسنقدر مساهماتك. ❤️`
  }
};

function setNoteLang(lang) {
  const trans = noteTranslations[lang];
  if (trans) {
    document.getElementById("noteTitleText").textContent = trans.title;
    document.getElementById("noteContentText").innerHTML = trans.content;

    const noteEl = document.getElementById("importantNote");
    if (lang === 'ar') {
      noteEl.style.direction = "rtl";
      noteEl.style.textAlign = "right";
    } else {
      noteEl.style.direction = "ltr";
      noteEl.style.textAlign = "left";
    }
  }
}