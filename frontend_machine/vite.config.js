import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";

function parseCSV(input, fallback = []) {
  const text = String(input || "").trim();
  if (!text) return fallback;
  return text.split(",").map((v) => v.trim()).filter(Boolean);
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  const port = Number(env.VITE_PORT || 5173);
  const apiTarget = env.VITE_API_TARGET || "http://127.0.0.1:5000";
  const wsTarget = env.VITE_WS_TARGET || apiTarget.replace(/^http/i, "ws");

  const allowedHosts = parseCSV(env.VITE_ALLOWED_HOSTS, [
    "localhost",
    "127.0.0.1",
    "10.5.0.8",
    "10.5.0.107",
  ]);

  return {
    plugins: [vue()],
    appType: "spa",
    server: {
      host: "0.0.0.0",
      port,
      strictPort: false,
      allowedHosts,
      proxy: {
        "/api": {
          target: apiTarget,
          changeOrigin: true,
          secure: false,
          timeout: 120000,
          proxyTimeout: 120000,
        },
        "/ws": {
          target: wsTarget,
          ws: true,
          changeOrigin: true,
          secure: false,
          timeout: 120000,
          proxyTimeout: 120000,
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
  };
});
