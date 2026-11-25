import { fileURLToPath } from "node:url";
import { Generator, getConfig } from "@tanstack/router-generator";
const root = fileURLToPath(new URL("../", import.meta.url));
await new Generator({
	root,
	config: getConfig(
		{
			routesDirectory: "./src/app",
			generatedRouteTree: "./src/routeTree.gen.ts",
			routeFileIgnorePattern: "\\.(test|stories)\\.",
			// Match the installed Start plugin's generated registration. Without this,
			// standalone codegen removes the footer that Vite restores on discovery.
			routeTreeFileFooter: [
				`import type { getRouter } from './router.tsx'
import type { createStart } from '@tanstack/react-start'
declare module '@tanstack/react-start' {
  interface Register {
    ssr: true
    router: Awaited<ReturnType<typeof getRouter>>
  }
}`,
			],
		},
		root,
	),
}).run();
