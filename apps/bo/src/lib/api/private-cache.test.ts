import { describe, expect, test } from "bun:test";
import { isStaffGovernanceQuery } from "./private-cache";

describe("staff private cache cleanup", () => {
	test("matches staff list, detail and sessions only", () => {
		for (const key of [
			["private", "staff-1", 3, "staff"],
			["private", "staff-1", 3, "staff-detail"],
			["private", "staff-1", 3, "staff-sessions"],
		]) expect(isStaffGovernanceQuery({ queryKey: key })).toBe(true);
		for (const key of [
			["private", "staff-1", 3, "customer-user", "user-1"],
			["private", "staff-1", 3, "organizations"],
			["identity", "staff-1", 3, "staff"],
		]) expect(isStaffGovernanceQuery({ queryKey: key })).toBe(false);
	});
});
