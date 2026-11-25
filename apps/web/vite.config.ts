import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite";
import { copyFile, mkdir } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

export default defineConfig({
	// Runtime config arrives through getPublicConfig, never Vite env injection.
	envDir: false,
	envPrefix: [],
	plugins: [
		tailwindcss(),
		// TanStack Start resolves router paths from the configured src directory.
		tanstackStart({ router: { routesDirectory: "./app", routeFileIgnorePattern: "\\.(test|stories)\\." } }),
		nitro({
			// Deployment adapter only, not a browser value or YAML substitution.
			preset: process.env.VERCEL === "1" ? "vercel" : "bun",
			// Register after the preset without replacing Vercel's artifact hook.
			modules: [
				(nitro) => {
					nitro.hooks.hook("compiled", async () => {
						const directory = join(nitro.options.output.serverDir, "config");
						await mkdir(directory, { recursive: true });
						await copyFile(fileURLToPath(new URL("../../config/config.yaml", import.meta.url)), join(directory, "config.yaml"));
					});
				},
			],
		}),
		react(),
	],
	resolve: { alias: { "@": new URL("./src", import.meta.url).pathname } },
});
