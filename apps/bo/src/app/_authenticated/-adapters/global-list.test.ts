import { describe, expect, test } from "bun:test";
import { listHandlers } from "./global-list";

const org = "00000000-0000-4000-8000-000000000101";
function recorder(current: Record<string, unknown>) {
	const calls: Record<string, unknown>[] = [];
	const navigate = (options: { search: (previous: Record<string, unknown>) => Record<string, unknown> }) => void calls.push(options.search(current));
	return { calls, handlers: listHandlers(navigate) };
}
describe("listHandlers", () => {
	test("an automatic filter change restarts at page 1 and keeps the organization filter", () => {
		const { calls, handlers } = recorder({ organizationId: org, q: "old", status: "failed", page: 4 });
		handlers.onFilter("baru", { status: "completed", stateGroup: "" });
		expect(calls[0]).toEqual({ organizationId: org, q: "baru", status: "completed", stateGroup: "", page: 1 });
	});
	test("clearing all filters restarts at page 1", () => {
		const { calls, handlers } = recorder({ q: "old", status: "failed", page: 3 });
		handlers.onFilter("", { status: "" });
		expect(calls[0]).toEqual({ q: "", status: "", page: 1 });
	});
	test("paging and the organization filter keep every other validated key", () => {
		const { calls, handlers } = recorder({ q: "x", page: 2 });
		handlers.onPage(3);
		handlers.onFilterOrganization(org);
		handlers.onClearOrganization();
		expect(calls).toEqual([
			{ q: "x", page: 3 },
			{ q: "x", organizationId: org, page: 1 },
			{ q: "x", organizationId: undefined, page: 1 },
		]);
	});
});
