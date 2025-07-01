import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: "/microscope",
  css: {
    modules: {
      localsConvention: "camelCase", // Converts kebab-case to camelCase for class names
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ["react", "react-dom"],
          mantine: [
            "@mantine/core",
            "@mantine/hooks",
            "@mantine/notifications",
          ],
        },
      },
    },
  },
  server: {
    port: 7000,
    proxy: {
      "/api": {
        target: "http://localhost:8089",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "/microscope/api"),
      },
    },
  },
});
