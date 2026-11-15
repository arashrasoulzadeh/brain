import path from "node:path";
import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Keep development hot-reload artifacts separate from production builds.
  // Sharing .next lets concurrent dev/build processes delete each other's chunks.
  distDir: process.env.NODE_ENV === "development" ? ".next-dev" : ".next",
  outputFileTracingRoot: path.join(__dirname, ".."),
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082"}/:path*`,
      },
      {
        source: "/public/:path*",
        destination: `${process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082"}/public/:path*`,
      },
    ];
  },
};
export default nextConfig;
