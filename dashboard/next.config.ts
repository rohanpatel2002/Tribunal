import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    root: __dirname,
  },
  output: "standalone",
  typescript: {
    ignoreBuildErrors: true,
  }
};

export default nextConfig;
