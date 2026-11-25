import { chromium, expect } from "@playwright/test";

// UI-only evidence against the ownership-checked repository Storybook. No server
// is started here, and no runtime configuration/API request is made.
const browser = await chromium.launch();
const page = await browser.newPage({ reducedMotion: "reduce" });
const errors: string[] = [];
page.on("pageerror", (error) => errors.push(error.message));
try {
	for (const width of [1280, 390]) {
		await page.setViewportSize({ width, height: 844 });
		for (const id of [
			"web-pages-foundation--default",
			"web-pages-foundation--narrow",
			"brand-identity--identity",
			"brand-identity--color-tokens",
		]) {
			await page.goto(`http://127.0.0.1:6006/iframe.html?id=${id}&viewMode=story`);
			await page.locator("#storybook-root").waitFor();
			if (id.startsWith("web-")) {
				await expect(page.getByRole("heading", { name: "Zenith", exact: true })).toBeVisible();
				await expect(page.getByRole("main")).toBeVisible();
				await expect(page.getByRole("button")).toHaveCount(0);
			} else if (id.endsWith("--identity")) {
				await expect(page.getByRole("img", { name: "Zenith logo" })).toBeVisible();
				await expect(page.getByText("Zenith logo and shared color tokens.")).toBeVisible();
			} else {
				await expect(page.getByText("--background", { exact: true })).toBeVisible();
			}
			expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
			await page.screenshot({ path: `/tmp/opencode/${id}-${width}.png` });
		}
	}
	expect(errors).toEqual([]);
	console.log(
		"PASS: 4 foundation/brand stories at 1280px and 390px; semantic content, no horizontal overflow or page errors. Static page has no controls.",
	);
} finally {
	await browser.close();
}
