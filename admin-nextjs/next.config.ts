import path from "node:path";
import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  reactStrictMode: true,
  outputFileTracingRoot: path.join(__dirname, ".."),
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082"}/:path*`,
      },
    ];
  },
};
export default nextConfig;
