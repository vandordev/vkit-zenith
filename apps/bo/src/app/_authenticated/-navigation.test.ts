import { describe, expect, test } from "bun:test";
import { resolveNavigation } from "./-navigation";

const org = "00000000-0000-4000-8000-000000000101";
const context = { kind: "organization", id: org, title: "Toko Maju Jaya" } as const;
describe("resolveNavigation", () => {
	test("global panel hides Staf and Log Aktivitas from admin", () => {
		const admin = resolveNavigation({ pathname: "/", search: {}, role: "admin" });
		const global = ["home", "organizations", "users", "jobs"];
		expect(admin.items.map((item) => item.id)).toEqual(global);
		// Divider before Operasional; none before the absent Sistem group.
		expect(admin.items.filter((item) => item.dividerAfter).map((item) => item.id)).toEqual(["users"]);
		expect(admin.panel).toEqual({ key: "global", depth: 0 });
		const superadmin = resolveNavigation({ pathname: "/staff", search: {}, role: "superadmin" });
		expect(superadmin.items.map((item) => item.id)).toEqual([...global, "staff", "activity"]);
		expect(superadmin.items.filter((item) => item.dividerAfter).map((item) => item.id)).toEqual(["users", "jobs"]);
		expect(superadmin.items.find((item) => item.id === "staff")?.active).toBe(true);
		expect(superadmin.sidebarContext).toBeUndefined();
	});
	test("organization panel lists its sections in order and marks the active one", () => {
		const resolved = resolveNavigation({ pathname: `/organizations/${org}/api-keys`, search: {}, role: "admin", context });
		expect(resolved.panel).toEqual({ key: "organization", depth: 1 });
		expect(resolved.items.map((item) => item.label)).toEqual([
			"Ringkasan",
			"Anggota",
			"API Key",
			"Log Aktivitas",
		]);
		expect(resolved.items.filter((item) => item.active).map((item) => item.id)).toEqual(["api-keys"]);
		expect(resolved.sidebarContext?.title).toBe("Toko Maju Jaya");
		expect(resolveNavigation({ pathname: `/organizations/${org}`, search: {}, role: "admin" }).items[0]?.active).toBe(true);
	});
	test("organization links carry only the organization id", () => {
		const resolved = resolveNavigation({ pathname: `/organizations/${org}/members`, search: {}, role: "admin", context });
		expect(resolved.items.every((item) => item.search === undefined && item.params?.organizationId === org)).toBe(true);
	});
	test("global links carry no search state and mark the active screen", () => {
		const resolved = resolveNavigation({ pathname: "/operations/jobs/1", search: {}, role: "superadmin" });
		expect(resolved.items.every((item) => item.search === undefined)).toBe(true);
		expect(resolved.items.filter((item) => item.active).map((item) => item.id)).toEqual(["jobs"]);
	});
	test("title falls back while the organization is still loading", () => {
		expect(resolveNavigation({ pathname: `/organizations/${org}`, search: {}, role: "admin" }).sidebarContext?.title).toBe("Organisasi");
		expect(
			resolveNavigation({ pathname: `/organizations/${org}`, search: {}, role: "admin", context: { ...context, id: "other" } })
				.sidebarContext?.title,
		).toBe("Organisasi");
	});
	test("user context panel lists its four sections, carries no search state and clears with the global panel", () => {
		const user = { kind: "user", id: org, title: "Nadia Prameswari" } as const;
		const resolved = resolveNavigation({ pathname: `/users/${org}/sessions`, search: {}, role: "admin", context: user });
		expect(resolved.panel).toEqual({ key: "user", depth: 1 });
		expect(resolved.items.map((item) => item.label)).toEqual(["Ringkasan", "Organisasi", "Sesi login", "Log Aktivitas"]);
		expect(resolved.items.filter((item) => item.active).map((item) => item.id)).toEqual(["sessions"]);
		expect(resolved.items.every((item) => item.params?.userId === org && item.search === undefined)).toBe(true);
		expect(resolved.sidebarContext?.title).toBe("Nadia Prameswari");
		expect(resolved.sidebarContext?.back).toEqual({ label: "Semua pengguna", to: "/users" });
		expect(resolveNavigation({ pathname: `/users/${org}`, search: {}, role: "admin", context: user }).items[0]?.active).toBe(true);
		expect(resolveNavigation({ pathname: `/users/${org}`, search: {}, role: "admin" }).sidebarContext?.title).toBe("Pengguna");
		expect(resolveNavigation({ pathname: `/users/${org}`, search: {}, role: "admin", context: { ...user, id: "other" } }).sidebarContext?.title).toBe("Pengguna");
		expect(resolveNavigation({ pathname: "/users", search: {}, role: "admin" }).panel.key).toBe("global");
	});
	test("user context preserves a validated membership origin on every tab and its back link", () => {
		const returnTo = `/organizations/${org}/members?q=nadia&status=active&ownership=owner&page=3&selected=${org}`;
		const resolved = resolveNavigation({ pathname: `/users/${org}/activity`, search: { customerReturnTo: returnTo }, role: "admin" });
		expect(resolved.items.every((item) => item.search?.returnTo === returnTo)).toBe(true);
		expect(resolved.sidebarContext?.back).toEqual({
			label: "Kembali ke daftar asal",
			to: "/organizations/$organizationId/members",
			params: { organizationId: org },
			search: { q: "nadia", status: "active", ownership: "owner", page: 3, selected: org },
		});
	});
	test("staff context panel is superadmin-only", () => {
		const staff = { kind: "staff", id: org, title: "Admin operasional" } as const;
		const resolved = resolveNavigation({ pathname: `/staff/${org}/activity`, search: {}, role: "superadmin", context: staff });
		expect(resolved.panel).toEqual({ key: "staff", depth: 1 });
		expect(resolved.items.map((item) => item.label)).toEqual(["Ringkasan", "Sesi", "Log Aktivitas"]);
		expect(resolved.items.filter((item) => item.active).map((item) => item.id)).toEqual(["activity"]);
		expect(resolved.sidebarContext?.back).toEqual({ label: "Semua staf", to: "/staff" });
		const admin = resolveNavigation({ pathname: `/staff/${org}`, search: {}, role: "admin", context: staff });
		expect(admin.panel.key).toBe("global");
		expect(admin.sidebarContext).toBeUndefined();
	});
	test("staff context carries only its validated list return through its tabs and back link", () => {
		const returnTo = "/staff?q=Nadia&status=disabled&page=4";
		const resolved = resolveNavigation({
			pathname: `/staff/${org}/sessions`,
			search: { returnTo },
			role: "superadmin",
			context: { kind: "staff", id: org, title: "Nadia Prameswari" },
		});
		expect(resolved.items.every((item) => item.search?.returnTo === returnTo)).toBe(true);
		expect(resolved.sidebarContext?.back).toEqual({
			label: "Semua staf",
			to: "/staff",
			search: { q: "Nadia", status: "disabled", page: 4 },
		});
	});
	test("organization list stays on the global panel", () => {
		expect(resolveNavigation({ pathname: "/organizations", search: {}, role: "admin" }).panel.key).toBe("global");
	});
});
