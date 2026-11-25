import { expect, test } from "bun:test";
import { activityRange } from "./activity-range";

const zones = {
	UTC: { spring: ["2026-03-08T00:00:00.000Z", "2026-03-09T00:00:00.000Z"], fall: ["2026-11-01T00:00:00.000Z", "2026-11-02T00:00:00.000Z"] },
	"Asia/Jakarta": {
		spring: ["2026-03-07T17:00:00.000Z", "2026-03-08T17:00:00.000Z"],
		fall: ["2026-10-31T17:00:00.000Z", "2026-11-01T17:00:00.000Z"],
	},
	"America/New_York": {
		spring: ["2026-03-08T05:00:00.000Z", "2026-03-09T04:00:00.000Z"],
		fall: ["2026-11-01T04:00:00.000Z", "2026-11-02T05:00:00.000Z"],
	},
};
function expected() {
	const zone = process.env.TZ;
	if (zone !== "UTC" && zone !== "Asia/Jakarta" && zone !== "America/New_York") throw new Error("Run with an explicit supported test TZ");
	return zones[zone];
}
test("spring range ends at next local calendar day, including 23-hour DST day", () => {
	const [from, until] = expected().spring;
	expect(activityRange("2026-03-08", "2026-03-08")).toEqual({ from, until });
});
test("fall range ends at next local calendar day, including 25-hour DST day", () => {
	const [from, until] = expected().fall;
	expect(activityRange("2026-11-01", "2026-11-01")).toEqual({ from, until });
});
test("clear and one-sided boundaries do not invent another bound", () => {
	const [from, until] = expected().spring;
	expect(activityRange()).toEqual({});
	expect(activityRange("", "")).toEqual({});
	expect(activityRange("2026-03-08")).toEqual({ from });
	expect(activityRange(undefined, "2026-03-08")).toEqual({ until });
});
test("reject malformed, overflowing and reversed calendar ranges safely", () => {
	for (const value of [
		"2026-02-29",
		"2024-02-30",
		"2026-04-31",
		"2026-00-01",
		"2026-13-01",
		"2026-01-00",
		"2026-3-08",
		"2026-03-08T00:00Z",
		"not a date",
		"0000-01-01",
	]) {
		expect(() => activityRange(value)).toThrow("Rentang tanggal tidak valid.");
		expect(() => activityRange(undefined, value)).toThrow("Rentang tanggal tidak valid.");
	}
	expect(() => activityRange("2026-11-02", "2026-11-01")).toThrow("Rentang tanggal tidak valid.");
});
test("leap day, month/year rollover and years below 100 use real calendar arithmetic", () => {
	for (const [date, next] of [
		["2024-02-29", "2024-03-01"],
		["2026-12-31", "2027-01-01"],
		["0099-12-31", "0100-01-01"],
	]) {
		const result = activityRange(date, date);
		const expectedEnd = new Date(0);
		const parts = next?.split("-").map(Number);
		if (!parts || parts[0] === undefined || parts[1] === undefined || parts[2] === undefined) throw new Error("Invalid fixture");
		expectedEnd.setFullYear(parts[0], parts[1] - 1, parts[2]);
		expectedEnd.setHours(0, 0, 0, 0);
		expect(result.until).toBe(expectedEnd.toISOString());
		expect(result.from).toBeDefined();
	}
});
