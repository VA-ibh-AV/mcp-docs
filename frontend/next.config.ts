import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  // Enable HMR in Docker with file polling
  // Using webpack mode explicitly for Docker compatibility
  webpack: (config, { dev, isServer }) => {
    if (dev && !isServer) {
      config.watchOptions = {
        poll: 1000, // Check for changes every second
        aggregateTimeout: 300, // Delay before rebuilding once the first file changed
      };
    }
    return config;
  },
  // Add empty turbopack config to silence the warning when using webpack
  turbopack: {},
};

export default nextConfig;
