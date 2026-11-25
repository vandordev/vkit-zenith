import { chromium, expect } from "@playwright/test";

// Explicit UI-only probe against an already-authorized Storybook; never starts services.
const origin = "http://127.0.0.1:6006";
const response = await fetch(`${origin}/index.json`);
if (!response.ok) throw new Error(`Storybook index returned ${response.status}`);
const index: { entries: Record<string, { id: string; type: string; title: string }> } = await response.json();
const titles = new Set(["UI/Input", "UI/Table", "Web/Components/HeroText"]);
const stories = Object.values(index.entries).filter((entry) => entry.type === "story" && titles.has(entry.title));
for (const title of titles) if (!stories.some((entry) => entry.title === title)) throw new Error(`Missing story: ${title}`);
const browser = await chromium.launch();
const page = await browser.newPage({ reducedMotion: "reduce" });
const errors: string[] = [];
page.on("pageerror", (error) => errors.push(error.message));
page.on("request", (request) => {
	if (new URL(request.url()).pathname.startsWith("/trpc")) errors.push("Story attempted a real API request");
});
async function open(id: string) {
	await page.goto(`${origin}/iframe.html?id=${id}&viewMode=story`);
	await page.locator("#storybook-root").waitFor();
	await expect(page.locator("#storybook-root")).not.toBeEmpty();
	await expect(page.locator(".sb-errordisplay")).not.toBeVisible();
}
try {
	for (const width of [1280, 390]) {
		await page.setViewportSize({ width, height: 844 });
		for (const story of stories) {
			await open(story.id);
			expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${story.id} overflow at ${width}`).toBe(true);
		}
	}
	await open("ui-input--form");
	const title = page.getByRole("textbox", { name: "Document title" });
	await page.keyboard.press("Tab");
	await expect(title).toBeFocused();
	expect(await title.evaluate((element) => getComputedStyle(element).boxShadow)).not.toBe("none");
	await title.fill("Updated local title");
	await page.getByRole("button", { name: "Save title", exact: true }).click();
	await expect(page.getByRole("status")).toHaveText("Saved: Updated local title");
	await page.getByRole("button", { name: "Focus title", exact: true }).click();
	await expect(title).toBeFocused();
	await open("ui-input--disabled");
	await expect(page.getByRole("textbox")).toBeDisabled();
	await open("ui-input--invalid");
	await expect(page.getByRole("textbox")).toHaveAttribute("aria-invalid", "true");
	await expect(page.getByRole("textbox")).toHaveAccessibleDescription("Enter your full name.");
	await open("ui-input--file");
	await page
		.getByLabel("Choose a .txt or .md document")
		.setInputFiles({ name: "notes.md", mimeType: "text/markdown", buffer: Buffer.from("# Notes") });
	await expect(page.getByRole("status")).toHaveText("notes.md");
	expect(errors).toEqual([]);
	console.log(
		`PASS: ${stories.length} generic primitive stories at desktop/mobile; form submit, ref/focus, disabled/invalid and file selection. No page errors/API requests.`,
	);
} finally {
	await browser.close();
}
