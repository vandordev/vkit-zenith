import { expect, test } from "bun:test";
import { orgLink, resolveCommandItems, resolveOrgNavigation } from "./navigation";

const resolve = (pathname: string) => resolveOrgNavigation({ pathname, slug: "kopi" });

test("links carry only the organization slug", () => {
	for (const pathname of ["/w/kopi", "/w/kopi/settings/profile"]) {
		const items = resolve(pathname).items;
		expect(items.length).toBeGreaterThan(0);
		for (const item of items) expect(item.params).toEqual({ organizationSlug: "kopi" });
	}
});

test("settings panel only under settings, with the Pengaturan item opening profile", () => {
	expect(resolve("/w/kopi").panel).toBe("main");
	expect(resolve("/w/kopi/settings/members").panel).toBe("settings");
	expect(resolve("/w/kopi").items.find((item) => item.id === "settings")?.to).toBe("/w/$organizationSlug/settings/profile");
	expect(
		resolve("/w/kopi")
			.items.filter((item) => item.active)
			.map((item) => item.id),
	).toEqual(["overview"]);
});

test("settings panel lists every settings menu in its group and marks the active one", () => {
	const items = resolve("/w/kopi/settings/api-keys").items;
	expect(items.map((item) => [item.id, item.group])).toEqual([
		["profile", "organization"],
		["members", "organization"],
		["api-keys", "developers"],
	]);
	expect(items.filter((item) => item.active).map((item) => item.id)).toEqual(["api-keys"]);
});

test("orgLink keeps extra params and search", () => {
	expect(orgLink("kopi", "/settings/api-keys", {}, { status: "revoked" })).toEqual({
		to: "/w/$organizationSlug/settings/api-keys",
		params: { organizationSlug: "kopi" },
		search: { status: "revoked" },
	});
});

test("command items list both panels once and other organizations only", () => {
	const items = resolveCommandItems({
		slug: "kopi",
		organizations: [
			{ slug: "kopi", name: "Kopi" },
			{ slug: "teh", name: "Teh" },
		],
	});
	const ids = items.map((item) => item.id);
	expect(new Set(ids).size).toBe(ids.length);
	expect(ids).not.toContain("settings");
	expect(ids).toContain("overview");
	expect(ids).toContain("api-keys");
	expect(ids.filter((id) => id.startsWith("organization-"))).toEqual(["organization-teh"]);
	expect(items.find((item) => item.id === "organization-teh")?.link.params).toEqual({ organizationSlug: "teh" });
});
