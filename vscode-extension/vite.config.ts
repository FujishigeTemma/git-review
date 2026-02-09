/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import { builtinModules } from "node:module";

export default defineConfig({
  resolve: {
    conditions: ["node"],
  },
  build: {
    lib: {
      entry: "src/extension.ts",
      formats: ["cjs"],
      fileName: "extension",
    },
    outDir: "dist",
    sourcemap: true,
    rollupOptions: {
      external: [
        "vscode",
        ...builtinModules,
        ...builtinModules.map((m) => `node:${m}`),
      ],
    },
    minify: false,
  },
  test: {
    includeSource: ["src/**/*.ts"],
  },
  define: {
    "import.meta.vitest": "undefined",
  },
});
