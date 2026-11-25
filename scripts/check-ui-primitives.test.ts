import { describe, expect, test } from "bun:test";
import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { checkUiPrimitives, inspectUiSource } from "./check-ui-primitives";

describe("shared UI primitive policy tooling", () => {
	for (const tag of [
		"button",
		"input",
		"select",
		"option",
		"optgroup",
		"textarea",
		"label",
		"dialog",
		"details",
		"summary",
		"progress",
		"meter",
		"table",
		"caption",
		"thead",
		"tbody",
		"tfoot",
		"tr",
		"th",
		"td",
	]) {
		test(`rejects native ${tag}`, () =>
			expect(inspectUiSource(`const view = <${tag} />;`, "legacy.stories.tsx")[0]).toContain("legacy.stories.tsx:1:"));
	}
	for (const source of [
		'const Control = "button"; const view = <Control />;',
		"const view = <motion.button />;",
		'createElement("input", {});',
		'const Tag = "input"; React.createElement(Tag);',
		'jsx("select", {}); jsxs("table", {});',
		'motion.create("button");',
		'import { Button } from "@/components/ui/button";',
		'export { Input } from "./input";',
		'import { Input } from "../../ui/input";',
		'import { Button } from "@base-ui/react/button";',
		'import * as Dialog from "@radix-ui/react-dialog";',
		'import { Slot } from "radix-ui/slot";',
	])
		test(`rejects common bypass: ${source}`, () => expect(inspectUiSource(source).length).toBeGreaterThan(0));
	for (const source of [
		'const view = <div className="font-display" />;',
		'const view = <div className="text-xl font-display tracking-tight" />;',
		'const view = <div className="font-[Instrument_Serif]" />;',
		'const view = <div className="font-serif" />;',
		"const view = <div className={`font-display ${x}`} />;",
	])
		test(`rejects serif display type: ${source}`, () => {
			const violations = inspectUiSource(source, "apps/bo/src/x.tsx");
			expect(violations.length).toBeGreaterThan(0);
			expect(violations[0]).toContain("serif display type was removed");
		});
	test("rejects serif utilities inside packages/ui as well", () => {
		expect(
			inspectUiSource('const view = <h1 className="font-display" />;', "packages/ui/src/components/ui/page-header.tsx").length,
		).toBeGreaterThan(0);
	});
	test("allows allied utilities that merely resemble the serif utility", () => {
		expect(inspectUiSource('const view = <div className="font-sans font-mono font-semibold" />;', "apps/bo/src/x.tsx")).toEqual([]);
	});
	test("allows shared primitives, structural markup, router links and SVG", () => {
		expect(
			inspectUiSource(
				'import { Button } from "@repo/ui/button"; import { Input } from "@repo/ui/input"; const view = <form><h1>Title</h1><labelled-section /><Input /><Button /><Link to="/" /><a href="/">Home</a><svg><path /></svg></form>;',
			),
		).toEqual([]);
	});
	test("scans all apps and stories, excluding tests, packages and generated outputs", () => {
		const root = mkdtempSync("/tmp/opencode/ui-primitives-");
		for (const directory of [
			"apps/web/src",
			"apps/landing/src",
			"apps/landing/.next",
			"apps/landing/out",
			"apps/storybook",
			"apps/web/dist",
			"apps/web/.output",
			"apps/web/node_modules",
			"packages/ui",
		])
			mkdirSync(join(root, directory), { recursive: true });
		for (const path of [
			"apps/web/src/legacy.tsx",
			"apps/landing/src/page.tsx",
			"apps/landing/.next/generated.js",
			"apps/landing/out/generated.js",
			"apps/storybook/fixture.stories.tsx",
			"apps/web/src/native.test.tsx",
			"apps/web/src/native.spec.ts",
			"apps/web/dist/generated.js",
			"apps/web/.output/generated.js",
			"apps/web/node_modules/vendor.tsx",
			"packages/ui/input.tsx",
		])
			writeFileSync(join(root, path), "const view = <input />;");
		expect(checkUiPrimitives(root)).toHaveLength(3);
	});
});
