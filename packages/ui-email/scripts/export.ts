import { render } from "@react-email/render";
import { brand } from "@repo/brand";
import { createElement } from "react";
import { createHash } from "node:crypto";
import { readFileSync, mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { AuthEmail, families, type EmailFields, type EmailPurpose, type EmailTheme } from "../src/templates";

const brandCSS = readFileSync(fileURLToPath(import.meta.resolve("@repo/brand/tokens.css")), "utf8");
const light = brandCSS.match(/:root\s*\{([^}]+)\}/)?.[1];
if (!light) throw new Error("brand light tokens unavailable");

// Email clients require broadly supported sRGB. This narrow export-only adapter
// accepts brand hex and achromatic OKLCH tokens; it never invents a second palette.
function color(name: string): string {
	const raw = light?.match(new RegExp(`--${name}:\\s*([^;]+);`))?.[1]?.trim();
	if (raw && /^#[0-9a-f]{6}$/i.test(raw)) return raw.toLowerCase();
	const neutral = raw?.match(/^oklch\(([\d.]+) 0 0\)$/);
	if (!neutral) throw new Error("unsupported email brand color");
	const l = Number(neutral[1]);
	if (l < 0 || l > 1) throw new Error("invalid email brand lightness");
	const linear = l ** 3;
	const encoded = linear <= 0.0031308 ? 12.92 * linear : 1.055 * linear ** (1 / 2.4) - 0.055;
	const channel = Math.round(encoded * 255)
		.toString(16)
		.padStart(2, "0");
	return `#${channel}${channel}${channel}`;
}

export const emailColors: EmailTheme = {
	background: color("background"),
	foreground: color("foreground"),
	primary: color("primary"),
	primaryForeground: color("primary-foreground"),
	mutedForeground: color("muted-foreground"),
};
const purposes = ["verify_email", "setup_password", "reset_password"] as const;
const actions: EmailFields = { DisplayName: "[[.DisplayName]]", ActionURL: "[[.ActionURL]]", ExpiresAt: "[[.ExpiresAt]]" };

function plain(purpose: EmailPurpose, fields: EmailFields) {
	const f = families[purpose];
	return `${brand.name}\n\n${f.heading}\n\nHalo ${fields.DisplayName},\n\n${f.instruction}\n\n${f.action}:\n${fields.ActionURL}\n\nTautan berlaku sampai ${fields.ExpiresAt}. Membuka tautan saja tidak menyelesaikan tindakan.\n\n${f.closing}\n`;
}

export async function exportTemplates() {
	async function one(purpose: EmailPurpose) {
		const html = await render(createElement(AuthEmail, { purpose, fields: actions, theme: emailColors }));
		const text = plain(purpose, actions);
		for (const field of Object.keys(actions)) {
			if (!html.includes(`[[.${field}]]`) || !text.includes(`[[.${field}]]`)) throw new Error("missing exported template action");
		}
		if (!html.includes('href="[[.ActionURL]]"') || html.includes("preview.example.test")) throw new Error("invalid exported dynamic link");
		return { subject: families[purpose].subject, html: `${html.trim()}\n`, text };
	}
	return {
		verify_email: await one("verify_email"),
		setup_password: await one("setup_password"),
		reset_password: await one("reset_password"),
	};
}

export async function generate() {
	const output = await exportTemplates();
	const templates = fileURLToPath(new URL("../../email/templates/", import.meta.url));
	const previews = fileURLToPath(new URL("../previews/", import.meta.url));
	mkdirSync(templates, { recursive: true });
	mkdirSync(previews, { recursive: true });
	const sources = [
		new URL("../src/templates.tsx", import.meta.url),
		new URL(import.meta.url),
		new URL("../../brand/src/tokens.css", import.meta.url),
		new URL("../../brand/src/index.ts", import.meta.url),
	];
	const sourceHash = createHash("sha256");
	for (const source of sources) sourceHash.update(readFileSync(source));
	for (const purpose of purposes) {
		writeFileSync(`${templates}/${purpose}.html.tmpl`, output[purpose].html);
		writeFileSync(`${templates}/${purpose}.txt.tmpl`, output[purpose].text);
		const fields: EmailFields = {
			DisplayName: "Pengguna Contoh",
			ActionURL: `https://preview.example.test/${purpose}?token=preview-only-not-a-credential`,
			ExpiresAt: "30 Oktober 2026, 12.00 UTC",
		};
		const previewHTML = await render(createElement(AuthEmail, { purpose, fields, theme: emailColors }));
		writeFileSync(`${previews}/${purpose}.html`, `${previewHTML}\n`);
		writeFileSync(`${previews}/${purpose}.mjs`, `export default ${JSON.stringify(previewHTML)};\n`);
	}
	writeFileSync(
		`${templates}/manifest.json`,
		JSON.stringify(
			{
				source: "packages/ui-email",
				sourceSHA256: sourceHash.digest("hex"),
				fields: Object.keys(actions),
				templates: Object.fromEntries(
					purposes.map((p) => [p, { subject: output[p].subject, html: `${p}.html.tmpl`, text: `${p}.txt.tmpl` }]),
				),
			},
			null,
			2,
		) + "\n",
	);
}

if (import.meta.main) await generate();
