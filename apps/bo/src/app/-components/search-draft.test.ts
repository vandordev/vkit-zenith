import { describe, expect, test } from "bun:test";
import { isOwnQueryEcho, searchLocationChanged, shouldResetSearchDraft } from "./search-draft";

describe("automatic search URL synchronization", () => {
	test("cancels a pending same-query draft when page or mode changes in history", () => {
		const current = { q: "nadia", navigationKey: JSON.stringify({ page: 3, mode: "test" }) };
		expect(shouldResetSearchDraft(current, { q: "nadia", navigationKey: JSON.stringify({ page: 2, mode: "test" }) }, true)).toBe(true);
		expect(shouldResetSearchDraft(current, { q: "nadia", navigationKey: JSON.stringify({ page: 3, mode: "live" }) }, true)).toBe(true);
	});
	test("opening or traversing the create drawer cancels a pending same-query draft", () => {
		const closed = { q: "nadia", navigationKey: JSON.stringify({ page: 3, create: undefined }) };
		const open = { q: "nadia", navigationKey: JSON.stringify({ page: 3, create: true }) };
		expect(shouldResetSearchDraft(closed, open, false)).toBe(true);
		expect(shouldResetSearchDraft(open, closed, false)).toBe(true);
		expect(shouldResetSearchDraft(closed, open, true)).toBe(true);
	});
	test("synchronizes changed query text and ignores an unchanged location", () => {
		const current = { q: "nadia", navigationKey: JSON.stringify({ page: 1 }) };
		expect(searchLocationChanged(current, { q: "nadi", navigationKey: current.navigationKey })).toBe(true);
		expect(searchLocationChanged(current, current)).toBe(false);
	});
	test("syncs history navigation while preserving a newer draft through its own URL echo", () => {
		const before = { q: "nadia", navigationKey: "page=3" };
		expect(shouldResetSearchDraft(before, { q: "nadia", navigationKey: "page=2" }, false)).toBe(true);
		expect(shouldResetSearchDraft(before, { q: "baru", navigationKey: "page=1" }, true)).toBe(false);
		expect(isOwnQueryEcho({ from: "nadia", to: "baru" }, before, { q: "baru", navigationKey: "page=1" })).toBe(true);
		expect(isOwnQueryEcho({ from: "nadia", to: "baru" }, { q: "lain", navigationKey: "page=2" }, { q: "baru", navigationKey: "page=3" })).toBe(false);
	});
});
