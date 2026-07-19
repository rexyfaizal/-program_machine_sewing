import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],

  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: false,

    allowedHosts: [
      "localhost",
      "127.0.0.1",
      "10.5.0.8",
    ],

    proxy: {
      "/api": {
        target: "http://127.0.0.1:5000",
        changeOrigin: true,
        secure: false,
        timeout: 30000,
        proxyTimeout: 30000,
      },

      "/ws": {
        target: "ws://127.0.0.1:5000",
        ws: true,
        changeOrigin: true,
        secure: false,
        timeout: 30000,
        proxyTimeout: 30000,

        configure: (proxy) => {
          proxy.on("error", (err) => {
            if (err.code === "ECONNABORTED" || err.code === "ECONNRESET") {
              console.warn("[vite proxy] WebSocket disconnected:", err.code);
              return;
            }

            console.error("[vite proxy] WebSocket proxy error:", err);
          });
        },
      },
    },
  },
});