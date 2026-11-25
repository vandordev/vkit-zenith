import { expect, test } from "bun:test";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { exportTemplates, emailColors } from "../scripts/export";

test("three auth email exports preserve Go actions and reproduce identical HTML/plain text", async () => {
	const first = await exportTemplates();
	const second = await exportTemplates();
	expect(first).toEqual(second);
	expect(Object.keys(first)).toEqual(["verify_email", "setup_password", "reset_password"]);
	for (const entry of Object.values(first)) {
		for (const field of ["DisplayName", "ActionURL", "ExpiresAt"]) {
			expect(entry.html).toContain(`[[.${field}]]`);
			expect(entry.text).toContain(`[[.${field}]]`);
		}
		expect(entry.html).toMatch(/href="\[\[\.ActionURL\]\]"/);
		expect(entry.html).not.toMatch(/preview\.example\.test|sample-recipient|https:\/\/react\.email/);
		expect(entry.text).not.toMatch(/<[^>]+>|preview\.example\.test/);
		expect(entry.html).not.toMatch(/<script|<img|<iframe|https?:\/\/(?:fonts|cdn)/i);
		expect(entry.html).toContain("Zenith");
		expect(entry.html).toContain(emailColors.primary);
	}
});

test("committed Go artifacts and provenance manifest exactly match current source", async () => {
	const root = fileURLToPath(new URL("../../email/templates/", import.meta.url));
	const exported = await exportTemplates();
	const manifest = JSON.parse(readFileSync(`${root}/manifest.json`, "utf8")) as {
		source: string;
		sourceSHA256: string;
		fields: string[];
		templates: Record<string, { subject: string; html: string; text: string }>;
	};
	expect(manifest.source).toBe("packages/ui-email");
	expect(manifest.fields).toEqual(["DisplayName", "ActionURL", "ExpiresAt"]);
	expect(Object.keys(exported)).toEqual(Object.keys(manifest.templates));
	for (const purpose of Object.keys(exported) as (keyof typeof exported)[]) {
		const item = exported[purpose];
		const entry = manifest.templates[purpose];
		expect(entry).toBeDefined();
		if (!entry) throw new Error("generated manifest entry missing");
		expect(entry.subject).toBe(item.subject);
		expect(readFileSync(`${root}/${entry.html}`, "utf8")).toBe(item.html);
		expect(readFileSync(`${root}/${entry.text}`, "utf8")).toBe(item.text);
	}
});

test("email color adapter consumes brand light tokens rather than a second palette", () => {
	expect(emailColors.primary).toBe("#1b1c1f");
	expect(emailColors.background).toMatch(/^#[0-9a-f]{6}$/);
	expect(emailColors.foreground).toMatch(/^#[0-9a-f]{6}$/);
	// Relative luminance remains readable after the neutral-only OKLCH adapter.
	expect(parseInt(emailColors.foreground.slice(1, 3), 16)).toBeLessThan(40);
	expect(parseInt(emailColors.background.slice(1, 3), 16)).toBeGreaterThan(240);
});
