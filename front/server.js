import { createRequestHandler } from "@react-router/express";
import { createProxyMiddleware } from "http-proxy-middleware";
import express from "express";

const app = express();

const BACKEND_URL = process.env.BACKEND_URL || "http://edge-metrics-server:8081";

app.use(
  "/api",
  createProxyMiddleware({
    target: BACKEND_URL,
    changeOrigin: true,
    pathRewrite: {
      "^/api": "",
    },
    onProxyReq: (_proxyReq, req) => {
      console.log(`[Proxy] ${req.method} ${req.url} -> ${BACKEND_URL}${req.url.replace(/^\/api/, "")}`);
    },
    onError: (err, req, res) => {
      console.error(`[Proxy Error] ${req.url}:`, err.message);
      res.status(502).json({ error: "Backend service unavailable" });
    },
  })
);

app.use(express.static("build/client", { maxAge: "1h" }));

app.all("*", createRequestHandler({ build: await import("./build/server/index.js") }));

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
  console.log(`Proxying /api requests to ${BACKEND_URL}`);
});
