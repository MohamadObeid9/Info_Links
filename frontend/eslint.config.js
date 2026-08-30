import js from "@eslint/js";
import globals from "globals";

/** Functions attached to window and called across modules / inline HTML. */
const windowGlobals = {
  setBtnLoading: "readonly",
  closeModal: "readonly",
  openModal: "readonly",
  renderProgTabs: "readonly",
  showView: "readonly",
  showToast: "readonly",
  apiRequest: "readonly",
  AppState: "readonly",
};

export default [
  {
    ignores: ["dist/**", "node_modules/**", "dev-dist/**"],
  },
  {
    files: ["js/**/*.js", "main.js"],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        ...globals.browser,
        ...windowGlobals,
      },
    },
    rules: {
      ...js.configs.recommended.rules,
      "no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrors: "none",
        },
      ],
      "no-empty": ["error", { allowEmptyCatch: true }],
    },
  },
  {
    files: ["js/**/*.test.js"],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.vitest,
      },
    },
  },
];
