import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api/order": {
        target: "http://127.0.0.1:8081",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/order/, "")
      },
      "/api/inventory": {
        target: "http://127.0.0.1:8082",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/inventory/, "")
      },
      "/api/query": {
        target: "http://127.0.0.1:8083",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/query/, "")
      },
      "/api/gateway": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/gateway/, "")
      },
      "/api/worker": {
        target: "http://127.0.0.1:8084",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/worker/, "")
      },
      "/api/nginx": {
        target: "http://127.0.0.1:8088",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/nginx/, "")
      }
    }
  }
});
