import { fileURLToPath } from "node:url";
import { Generator, getConfig } from "@tanstack/router-generator";

const root = fileURLToPath(new URL("../", import.meta.url));
const config = getConfig({
	routesDirectory: "./src/app",
	generatedRouteTree: "./src/routeTree.gen.ts",
	routeFileIgnorePattern: "\\.(test|stories)\\.",
}, root);
await new Generator({ config, root }).run();
