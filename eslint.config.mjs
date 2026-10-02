import js from "@eslint/js";
import globals from "globals";

export default [
  {
    ignores: ["node_modules/**", "test-artifacts/**", ".tidb-browser-profile/**"],
  },
  js.configs.recommended,
  {
    files: ["web/js/**/*.js"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: globals.browser,
    },
  },
  {
    files: ["scripts/**/*.js"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "commonjs",
      globals: globals.node,
    },
  },
];
