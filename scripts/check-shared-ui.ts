import { chromium, expect } from "@playwright/test";
import { mkdir, writeFile } from "node:fs/promises";

// UI-only probe. Uses the ownership-managed Storybook; never starts any server.
const origin = "http://127.0.0.1:6006";
const artifactDirectory = "/tmp/opencode/ui-verification";
const response = await fetch(`${origin}/index.json`);
if (!response.ok) throw new Error(`Storybook index returned ${response.status}`);
const index: { entries: Record<string, { id: string; type: string; title: string }> } = await response.json();
const stories = Object.values(index.entries).filter((entry) => entry.type === "story");
const titles = new Set(stories.filter((entry) => entry.title.startsWith("UI/")).map((entry) => entry.title));
expect(titles.size, "36 catalog components plus the local Table").toBe(37);
await mkdir(artifactDirectory, { recursive: true });
const browser = await chromium.launch();
const failures: { id: string; width: number; error: string }[] = [];
let checked = 0;

try {
	await Promise.all(
		[1280, 390].map(async (width) => {
			const page = await browser.newPage({ viewport: { width, height: 844 }, reducedMotion: "reduce", colorScheme: "dark" });
			page.setDefaultTimeout(10000);
			let errors: string[] = [];
			page.on("pageerror", (error) => errors.push(error.message));
			await page.route("**/*", async (route) => {
				const url = new URL(route.request().url());
				if (["http:", "https:"].includes(url.protocol) && url.origin !== origin) {
					errors.push(`Blocked external fixture request: ${url.origin}${url.pathname}`);
					await route.abort();
				} else if (/^\/(v1|trpc)(\/|$)/.test(url.pathname)) {
					errors.push(`Blocked business API request: ${url.pathname}`);
					await route.abort();
				} else {
					await route.continue();
				}
			});
			const photographed = new Set<string>();
			for (const story of stories) {
				errors = [];
				try {
					await page.goto(`${origin}/iframe.html?id=${story.id}&viewMode=story`);
					// Text emptiness incorrectly rejects SVG-only avatars and empty inputs.
					await expect(page.locator("#storybook-root > *").first()).toBeAttached();
					await expect(page.locator(".sb-errordisplay")).not.toBeVisible();
					await page.evaluate(() => document.fonts.ready);
					expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${story.id} overflow at ${width}`).toBe(
						true,
					);
					expect(errors, `${story.id} browser errors`).toEqual([]);
					if (!photographed.has(story.title)) {
						await page.screenshot({ path: `${artifactDirectory}/${story.id}-${width}.png` });
						photographed.add(story.title);
					}
					checked += 1;
				} catch (error) {
					failures.push({ id: story.id, width, error: error instanceof Error ? error.message : String(error) });
					await page.screenshot({ path: `${artifactDirectory}/failure-${story.id}-${width}.png` }).catch(() => undefined);
				}
			}
			await page.close();
		}),
	);
} finally {
	await browser.close();
	await writeFile(
		`${artifactDirectory}/summary.json`,
		JSON.stringify({ stories: stories.length, groups: titles.size, checked, failures }, null, 2),
	);
}
if (failures.length) {
	console.error(JSON.stringify(failures, null, 2));
	throw new Error(`${failures.length} failed story/viewport checks; ${checked} passed. See ${artifactDirectory}.`);
}
console.log(
	`PASS: ${stories.length} stories at 1280px and 390px, ${titles.size} UI groups; no page errors, external/API calls, or document overflow. Screenshots: ${artifactDirectory}`,
);
