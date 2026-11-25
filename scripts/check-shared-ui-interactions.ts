import { chromium, expect } from "@playwright/test";

// Local presentation fixtures only; never starts a service or calls a backend.
const origin = "http://127.0.0.1:6006";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 844 }, reducedMotion: "reduce" });
const errors: string[] = [];
page.on("pageerror", (error) => errors.push(error.message));
await page.route("**/*", async (route) => {
	const url = new URL(route.request().url());
	if (["http:", "https:"].includes(url.protocol) && url.origin !== origin) {
		errors.push(`Blocked external request: ${url.origin}${url.pathname}`);
		await route.abort();
	} else {
		await route.continue();
	}
});
const open = async (id: string) => {
	await page.goto(`${origin}/iframe.html?id=ui-${id}&viewMode=story`);
	await expect(page.locator("#storybook-root")).not.toBeEmpty();
};
const fits = async () => expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);

try {
	for (const width of [1280, 390]) {
		await page.setViewportSize({ width, height: 844 });
		for (const [component, role] of [
			["checkbox", "checkbox"],
			["switch", "switch"],
		] as const) {
			await open(`${component}--playground`);
			const control = page.getByRole(role);
			await page.keyboard.press("Tab");
			await expect(control).toBeFocused();
			await page.keyboard.press("Space");
			await expect(control).toBeChecked();
		}
		await open("accordion--playground");
		await page.getByRole("button", { name: "Recovery options" }).click();
		await expect(page.getByText("Check your connection and try again.")).toBeVisible();
		await open("tabs--playground");
		await page.getByRole("tab", { name: "Activity" }).click();
		await expect(page.getByRole("tabpanel")).toHaveText("Recent updates and conversations.");
		await open("select--controlled");
		await page.getByRole("combobox", { name: "Fruit", exact: true }).click();
		await page.getByRole("option", { name: "Banana" }).click();
		await expect(page.getByRole("status")).toHaveText("Selected: banana");
		await open("autocomplete--controlled");
		const autocomplete = page.getByRole("combobox");
		await autocomplete.fill("zzzz-no-match");
		await expect(page.getByText("No results found.", { exact: true })).toBeVisible();
		await page.keyboard.press("Escape");
		for (const [component, triggerName, dialogName] of [
			["dialog", "Open dialog", "Project settings"],
			["drawer", "Open drawer", "Project settings"],
		]) {
			await open(`${component}--playground`);
			const trigger = page.getByRole("button", { name: triggerName, exact: true });
			await trigger.click();
			await expect(page.getByRole("dialog", { name: dialogName })).toBeVisible();
			await fits();
			await page.keyboard.press("Escape");
			await expect(page.getByRole("dialog")).not.toBeVisible();
			await expect(trigger).toBeFocused();
		}
		await open("popover--playground");
		await page.getByRole("button", { name: "Manage access" }).click();
		await expect(page.getByText("Only invited members can view this workspace.")).toBeVisible();
		await page.getByRole("button", { name: "Done", exact: true }).click();
		await open("tooltip--playground");
		await page.keyboard.press("Tab");
		await expect(page.getByRole("button", { name: "Save changes" })).toBeFocused();
		await expect(page.locator('[data-slot="tooltip-content"]')).toHaveText("Save changes");
		await open("input-password--playground");
		await expect(page.locator("input")).toHaveAttribute("type", "password");
		await page.getByRole("button", { name: "Show password" }).click();
		await expect(page.locator("input")).toHaveAttribute("type", "text");
		await open("input-secret--empty");
		await page.getByRole("button", { name: "Copy demo secret" }).click();
		await expect(page.getByRole("status")).toHaveText("Secret is empty.");
		await open("input-amount--controlled");
		await page.getByRole("textbox").fill("1234.56");
		await expect(page.getByRole("status")).toHaveText("Decimal value: 1234.56");
		await open("input-phone--controlled");
		await page.getByRole("textbox", { name: "Phone number" }).fill("081234567890");
		await expect(page.getByRole("status")).toHaveText("International digits: 6281234567890");
		await open("datepicker--populated");
		await page.getByRole("button", { name: /Appointment date/ }).click();
		await expect(page.getByRole("grid")).toBeVisible();
		await fits();
		await page.getByRole("button", { name: "Clear date" }).click();
		await expect(page.getByRole("button", { name: /Appointment date/ })).toContainText("Select a date");
		await open("daterangepicker--shortcuts");
		await page.getByRole("button", { name: /Reporting period/ }).click();
		await expect(page.getByRole("grid").first()).toBeVisible();
		await fits();
		await page.keyboard.press("Escape");
		await open("toast--playground");
		await page.getByRole("button", { name: "Show configured notification" }).click();
		await expect(page.getByText("Changes saved", { exact: true }).first()).toBeVisible();
		await fits();
		await open("datagrid--playground");
		await expect(page.getByRole("cell", { name: "Example user 1", exact: true })).toBeVisible();
		await page.getByRole("searchbox", { name: "Search users" }).fill("Example user 60");
		await expect(page.getByRole("cell", { name: "Example user 60", exact: true })).toBeVisible();
		await expect(page.getByRole("cell", { name: "Example user 1", exact: true })).toHaveCount(0);
		await fits();
		await open("appshell--default");
		if (width < 768) {
			await page.getByRole("button", { name: "Open navigation" }).click();
			await expect(page.getByRole("dialog")).toBeVisible();
			await expect(page.getByRole("button", { name: "Overview", exact: true })).toBeVisible();
			await fits();
			await page.keyboard.press("Escape");
			await expect(page.getByRole("button", { name: "Open navigation" })).toBeFocused();
		} else {
			await page.getByRole("button", { name: "Collapse sidebar" }).click();
			await expect(page.getByRole("button", { name: "Expand sidebar" })).toBeVisible();
		}
	}
	expect(errors).toEqual([]);
	console.log(
		"PASS: desktop/mobile keyboard toggles, tabs, autocomplete, select, dialog/drawer focus restoration, popover, tooltip, password/secret feedback, amount/phone values, date overlays, toast, local DataGrid filtering, and AppShell navigation.",
	);
} finally {
	await browser.close();
}
