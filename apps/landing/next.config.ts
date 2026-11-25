import type { NextConfig } from "next";

const config: NextConfig = {
	output: "export",
	transpilePackages: ["@repo/ui", "@repo/brand"],
};

export default config;
