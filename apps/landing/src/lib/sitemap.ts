import { readdirSync } from "node:fs";
import { join } from "node:path";
import { isIndexablePath } from "./seo";

/** Discover static App Router pages without importing their components. */
export function sitemapPaths(directory: string, segments: string[] = []): string[] {
	const paths: string[] = [];
	for (const entry of readdirSync(directory, { withFileTypes: true })) {
		if (entry.isDirectory()) {
			// shortcut: parameterized routes need explicit generated URLs when they are introduced.
			if (/^[_@\[]/.test(entry.name)) continue;
			const next = entry.name.startsWith("(") ? segments : [...segments, entry.name];
			paths.push(...sitemapPaths(join(directory, entry.name), next));
		} else if (/^page\.(tsx?|jsx?)$/.test(entry.name)) {
			const path = `/${segments.join("/")}`;
			if (isIndexablePath(path)) paths.push(path);
		}
	}
	return [...new Set(paths)].sort();
}
