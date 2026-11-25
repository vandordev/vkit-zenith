import { expect, test } from "bun:test";
import { activityTarget, parseActivitySearch, isActivityQuery, activityFilterSearch } from "./activity-navigation";
import type { V1ActivityRow } from "./generated/types.gen";
const organization = "00000000-0000-4000-8000-000000000001",
	target = "00000000-0000-4000-8000-000000000002",
	actor = "00000000-0000-4000-8000-000000000003";
const row: V1ActivityRow = {
	id: target,
	occurredAt: "2026-10-26T00:00:00Z",
	action: "staff.update",
	outcome: "success",
	actor: { kind: "system", id: null, name: null },
	target: { type: "staff", id: target, organizationId: null, name: null, available: true },
};
test("customer targets use exact authoritative IDs and membership organization", () => {
	const customerRow = {
		...row,
		action: "customer.membership.role",
		target: { ...row.target, organizationId: organization },
	};
	expect(activityTarget({ ...customerRow, target: { ...customerRow.target, type: "user" } })).toBe(`/users/${target}`);
	expect(activityTarget({ ...customerRow, target: { ...customerRow.target, type: "organization" } })).toBe(`/organizations/${target}`);
	expect(activityTarget({ ...customerRow, target: { ...customerRow.target, type: "organization_membership" } })).toBe(
		`/organizations/${organization}/members?selected=${target}`,
	);
	for (const organizationId of [null, "bad", "//evil.invalid"])
		expect(
			activityTarget({ ...customerRow, target: { ...customerRow.target, type: "organization_membership", organizationId } }),
		).toBeUndefined();
	for (const targetChange of [{ available: false }, { id: null }, { id: "bad" }, { type: "future-customer" }])
		expect(activityTarget({ ...customerRow, target: { ...customerRow.target, type: "user", ...targetChange } })).toBeUndefined();
	expect(activityTarget({ ...row, target: { ...row.target, type: "organization_membership", organizationId: null } })).toBeUndefined();
});
const filters = {
	from: "2026-10-05",
	to: "2026-10-12",
	actorStaffId: actor,
	action: "staff.update",
	outcome: "success" as const,
};
for (const [name, patch] of Object.entries({
	from: { from: "2026-10-06" },
	to: { to: "2026-10-13" },
	dateRange: { from: "2026-10-06", to: "2026-10-13" },
	dateClear: { from: undefined, to: undefined },
	action: { action: "staff.login" },
	actionClear: { action: undefined },
	outcome: { outcome: "failure" as const },
	outcomeClear: { outcome: undefined },
	staff: { actorStaffId: target },
	staffClear: { actorStaffId: undefined },
})) {
	test(`filter transition ${name} whitelists search payload and retains selection`, () => {
		const current = parseActivitySearch({ ...filters, page: 3, selected: target, invalid: true });
		const nextFilters = { ...filters, ...patch };
		const payload = { ...current, ...patch, returnTo: "/staff", pageSize: 100, extra: "unwanted" };
		const next = activityFilterSearch(payload);
		expect(next).toEqual({ ...nextFilters, selected: target, page: 1, invalid: false });
		expect(next).not.toHaveProperty("returnTo");
		expect(next).not.toHaveProperty("pageSize");
		expect(next).not.toHaveProperty("extra");
		expect(parseActivitySearch(next)).toEqual({ ...nextFilters, selected: target, page: 1, invalid: false });
	});
}
test("filter reset removes filters and resets pagination", () => {
	const reset = activityFilterSearch({});
	expect(JSON.parse(JSON.stringify(reset))).toEqual({ page: 1, invalid: false });
	expect(reset).not.toHaveProperty("selected");
	expect(parseActivitySearch(reset)).toEqual({ ...reset, page: 1, invalid: false });
});
test("filter reset retains a valid open detail selection when supplied", () => {
	const reset = activityFilterSearch({ selected: target });
	expect(reset).toEqual({ selected: target, page: 1, invalid: false });
});
test("unavailable, internal, unknown and malformed identifiers have no action", () => {
	for (const changed of [
		{ type: "auth_summary" },
		{ type: "auth_bucket" },
		{ type: "future" },
		{ id: "//evil.invalid" },
		{ id: null },
		{ available: false },
	])
		expect(activityTarget({ ...row, target: { ...row.target, ...changed } })).toBeUndefined();
});
test("invalid URL state is explicit and never silently broadens a request", () => {
	expect(parseActivitySearch({ page: "3", selected: target, outcome: "denied" })).toEqual({
		page: 3,
		selected: target,
		outcome: "denied",
		invalid: false,
	});
	for (const value of [
		{ page: -1 },
		{ page: 1000001 },
		{ selected: "bad" },
		{ actorStaffId: "bad" },
		{ outcome: "invented" },
		{ from: "2026-10-26T00:00Z" },
	])
		expect(parseActivitySearch(value).invalid).toBe(true);
});
test("role loss evicts only identity-prefixed activity list/detail/choice/recent caches", () => {
	for (const key of ["activity", "activity-detail", "activity-actions", "activity-staff-choices", "activity-recent"])
		expect(isActivityQuery({ queryKey: ["private", target, 3, key, { page: 1 }] })).toBe(true);
	for (const queryKey of [["identity", 3], ["private", target, 3, "staff"], ["private", target, 3, "overview"], ["activity"], []])
		expect(isActivityQuery({ queryKey })).toBe(false);
});
