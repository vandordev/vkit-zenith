import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite";
import { copyFile, mkdir } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
export default defineConfig({
	envDir: false,
	envPrefix: [],
	plugins: [
		tailwindcss(),
		tanstackStart({ router: { routesDirectory: "./app", routeFileIgnorePattern: "\\.(test|stories)\\." } }),
		nitro({
			// Deployment adapter only; runtime API configuration is never baked in.
			preset: process.env.VERCEL === "1" ? "vercel" : "bun",
			// Register an additional hook without replacing the preset's compiled hook,
			// which generates Vercel's config.json and function metadata.
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
	resolve: { alias: { "@bo": fileURLToPath(new URL("./src", import.meta.url)) } },
});
