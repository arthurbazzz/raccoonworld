import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Self-contained server in .next/standalone for the Docker image (RW-007).
  output: "standalone",
};

export default nextConfig;
