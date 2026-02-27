import path from "node:path";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
  plugins: [solidPlugin()],
  envDir: path.resolve(__dirname, "../.."),
  server: {
    port: 3000,
  },
  build: {
    target: "esnext",
  },
});
