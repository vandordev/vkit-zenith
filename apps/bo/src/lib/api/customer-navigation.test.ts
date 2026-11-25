import { describe, expect, test } from "bun:test";
import {
	parseCustomerReturn,
	organizationSearch,
	organizationMembersSearch,
	userSearch,
	userDetailSearch,
	organizationOverviewSearch,
	organizationSectionSearch,
	customerReturn,
	customerOrigin,
	staffDetailSearch,
	staffListReturn,
	staffReturnSearch,
	parseStaffReturn,
} from "./customer-navigation";
const id = "00000000-0000-4000-8000-000000000101";
describe("customer navigation", () => {
	test("cross-context origin stays bounded and form secrets never enter URL", () => {
		const origin = "/users?q=Nadia&page=2";
		expect(customerOrigin(`/organizations/${id}`, { returnTo: origin, q: "member", page: 3 })).toBe(origin);
		expect(customerReturn("/users", { page: 1, password: "fixture-secret", reason: "secret" })).toBe("/users");
		expect(parseCustomerReturn("/users?currentPassword=fixture-secret")).toBeUndefined();
	});
	test("preserves known filters/page and bounded origin, never arbitrary return chains", () => {
		expect(parseCustomerReturn("/organizations?q=Koperasi%20Utama&status=suspended&page=3")).toBe(
			"/organizations?q=Koperasi+Utama&status=suspended&page=3",
		);
		const origin = "/users?q=Nadia&status=disabled&verification=verified&page=2";
		const current = customerReturn(`/organizations/${id}/members`, {
			q: "",
			status: "removed",
			ownership: "owner",
			page: 2,
			selected: id,
			returnTo: origin,
		});
		expect(parseCustomerReturn(current)).toBe(current);
		expect(new URL(current, "https://test.invalid").searchParams.get("returnTo")).toBe(origin);
	});
	test("organization context sub-routes keep their page and nothing else", () => {
		for (const section of ["/members", "/api-keys", "/activity"])
			expect(parseCustomerReturn(`/organizations/${id}${section}?page=2`)).toBe(`/organizations/${id}${section}?page=2`);
		expect(parseCustomerReturn(`/organizations/${id}`)).toBe(`/organizations/${id}`);
		expect(parseCustomerReturn(`/organizations/${id}?page=2`)).toBeUndefined();
		expect(customerReturn(`/organizations/${id}/members`, { page: 2 })).toBe(`/organizations/${id}/members?page=2`);
		for (const value of [
			`/organizations/${id}/unknown`,
			`/organizations/${id}/products/${id}`,
			`/organizations/${id}/api-keys?mode=test`,
			`/organizations/${id}/api-keys?selected=${id}`,
			`/organizations/${id}/members/${id}`,
			`/users?mode=test`,
			`/users/${id}/unknown`,
			`/users/${id}/sessions/${id}`,
			`/users/${id}/organizations?mode=test`,
			`/users/${id}/sessions?q=ab`,
			`/staff/${id}`,
		])
			expect(parseCustomerReturn(value)).toBeUndefined();
	});
	test("user context sections are valid origins with page and one validated return only", () => {
		for (const section of ["organizations", "sessions", "activity"])
			expect(parseCustomerReturn(`/users/${id}/${section}?page=2`)).toBe(`/users/${id}/${section}?page=2`);
		const origin = `/organizations/${id}/members?q=Nadia&status=active&ownership=owner&page=3&selected=${id}`;
		for (const section of ["organizations", "sessions", "activity"])
			expect(parseCustomerReturn(`/users/${id}/${section}?returnTo=${encodeURIComponent(origin)}`)).toBe(
				`/users/${id}/${section}?returnTo=${encodeURIComponent(origin)}`,
			);
		expect(parseCustomerReturn(`/users/${id}/organizations?returnTo=${encodeURIComponent("/users?page=3")}`)).toBe(
			`/users/${id}/organizations?returnTo=%2Fusers%3Fpage%3D3`,
		);
		expect(customerReturn(`/users/${id}/organizations`, { page: 2 })).toBe(`/users/${id}/organizations?page=2`);
	});
	test("user activity query and a one-level return origin remain valid", () => {
		const activity = `/users/${id}/activity?q=organization.updated&page=3`;
		expect(parseCustomerReturn(activity)).toBe(activity);
		expect(userDetailSearch.parse({ returnTo: activity }).returnTo).toBe(activity);
		expect(parseCustomerReturn(`/users/${id}/sessions?q=organization.updated&page=3`)).toBeUndefined();
		const membershipOrigin = `/organizations/${id}/members?q=Nadia&status=active&ownership=owner&page=3&selected=${id}`;
		const activityWithOrigin = `/users/${id}/activity?q=organization.updated&page=2&returnTo=${encodeURIComponent(membershipOrigin)}`;
		expect(parseCustomerReturn(activityWithOrigin)).toBe(activityWithOrigin);
		expect(customerOrigin(`/organizations/${id}/members`, { returnTo: activityWithOrigin, selected: id, page: 1 })).toBe(activityWithOrigin);
		expect(
			parseCustomerReturn(`/users/${id}/activity?q=again&returnTo=${encodeURIComponent(activityWithOrigin)}`),
		).toBeUndefined();
	});
	test("staff detail returns to its validated filtered and paged list", () => {
		const list = staffListReturn({ q: "Nadia Prameswari", status: "disabled", page: 4 });
		expect(list).toBe("/staff?q=Nadia+Prameswari&status=disabled&page=4");
		expect(parseStaffReturn(list)).toBe(list);
		expect(staffDetailSearch.parse({ returnTo: list }).returnTo).toBe(list);
		expect(staffReturnSearch(list)).toEqual({ q: "Nadia Prameswari", status: "disabled", page: 4 });
		for (const value of ["https://evil.test", "//evil.test/staff", "/staff?password=secret", "/staff?status=admin", "/staff?page=0", "/staff?returnTo=%2Fstaff"])
			expect(parseStaffReturn(value)).toBeUndefined();
		expect(staffReturnSearch(undefined)).toEqual({ q: "", status: "", page: 1 });
	});
	test("rejects unsafe paths, duplicates, secrets, invalid values and nested arbitrary URLs", () => {
		for (const value of [
			undefined,
			{},
			"https://evil.test/users",
			"//evil.test/users",
			"/\\evil.test/users",
			"/users/../../staff",
			"/users#secret",
			"/users?password=x",
			"/users?reason=secret",
			"/users?email=secret@example.test",
			"/users?page=1000001",
			"/users?page=-1",
			"/users?status=suspended",
			"/organizations?status=disabled",
			"/users?q=a&q=b",
			"/users?q=%00",
			"/users?create=true",
			`/organizations/${id}?returnTo=${encodeURIComponent("https://evil.test")}`,
			`/organizations/${id}?returnTo=${encodeURIComponent(`/users/${id}?returnTo=${encodeURIComponent("/organizations")}`)}`,
		])
			expect(parseCustomerReturn(value)).toBeUndefined();
	});
	test("route search defaults are bounded, UUID selection valid and origin safe", () => {
		expect(organizationMembersSearch.parse({ action: "owner" }).action).toBeUndefined();
		expect(organizationOverviewSearch.parse({ action: "add" })).toMatchObject({ action: undefined });
		expect(organizationSectionSearch.parse({})).toEqual({ page: 1 });
		expect(organizationSectionSearch.parse({ page: "3" })).toEqual({ page: 3 });
		expect(organizationSectionSearch.parse({ page: "0" }).page).toBe(1);
		expect(organizationSearch.parse({ page: "2", status: "suspended" })).toMatchObject({ q: "", page: 2, status: "suspended" });
		expect(organizationMembersSearch.parse({ selected: "bad", action: "secret", returnTo: "//evil.test" })).toMatchObject({
			selected: undefined,
			action: undefined,
			returnTo: undefined,
		});
		expect(userSearch.parse({ verification: "wrong", page: 0 })).toMatchObject({ verification: "", page: 1 });
		expect(userDetailSearch.parse({ action: "verify", returnTo: "/users?page=2" })).toMatchObject({
			action: "verify",
			returnTo: "/users?page=2",
		});
	});

});
