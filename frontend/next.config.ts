import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone", // required for Railway/Docker deployment
};

export default nextConfig;
