import { describe, expect, test } from "bun:test";
import { counterTarget } from "./home-counters";

describe("counterTarget", () => {
	test("the failed-jobs counter links to the job list filtered by state group", () => {
		expect(counterTarget("/operations/jobs?stateGroup=failed")).toEqual({
			to: "/operations/jobs",
			search: { q: "", page: 1, status: "", stateGroup: "failed" },
		});
	});
	test("a value the destination validator would change gets no link", () => {
		expect(counterTarget("/operations/jobs?stateGroup=exploded")).toBeUndefined();
	});
	test("unknown paths and other origins get no link", () => {
		expect(counterTarget("/organizations?status=active")).toBeUndefined();
		expect(counterTarget("https://evil.example/operations/jobs?stateGroup=failed")).toBeUndefined();
	});
});
