import path from "node:path";
import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  reactStrictMode: true,
  basePath: "/_system/ui",
  // Keep development hot-reload artifacts separate from production builds.
  // Sharing .next lets concurrent dev/build processes delete each other's chunks.
  distDir: process.env.NODE_ENV === "development" ? ".next-dev" : ".next",
  outputFileTracingRoot: path.join(__dirname, ".."),
  async rewrites() {
    const origin = process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082";
    return [
      { source: "/_system/admin/:path*", destination: `${origin}/_system/admin/:path*`, basePath: false },
      { source: "/:project/admin/:path*", destination: `${origin}/:project/admin/:path*`, basePath: false },
      { source: "/:project/functions/:name", destination: `${origin}/:project/functions/:name`, basePath: false },
      { source: "/:project([A-Za-z0-9][A-Za-z0-9_-]*)/:resource*", destination: `${origin}/:project/:resource*`, basePath: false },
    ];
  },
};
export default nextConfig;
