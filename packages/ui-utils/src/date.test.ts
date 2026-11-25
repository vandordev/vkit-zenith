import { describe, expect, test } from "bun:test";
import { formatDate, toDate } from "./date";

describe("date presentation", () => {
	test("converts UTC instants into the chosen local zone with the default format", () => {
		expect(formatDate("2026-10-24T08:00:00Z", { timeZone: "Asia/Jakarta" })).toBe("24 Okt 2026, 15.00");
	});
	test("supports format and locale overrides without conflicting with defaults", () => {
		expect(formatDate("2026-10-24T08:00:00Z", { locale: "en-GB", timeZone: "UTC", format: { dateStyle: "long" } })).toBe("24 October 2026");
	});
	test("handles day rollover and fractional-offset zones", () => {
		expect(formatDate("2026-10-24T23:30:00Z", { timeZone: "Asia/Kathmandu" })).toBe("25 Okt 2026, 05.15");
	});
	test("uses daylight-saving offsets for the date, not the current date", () => {
		expect(formatDate("2026-03-08T06:30:00Z", { timeZone: "America/New_York", locale: "en-GB" })).toBe("8 Mar 2026, 01:30");
		expect(formatDate("2026-03-08T07:30:00Z", { timeZone: "America/New_York", locale: "en-GB" })).toBe("8 Mar 2026, 03:30");
	});
	test("accepts explicit offsets, Go fractional seconds, Date objects and epoch milliseconds", () => {
		const expected = "24 Okt 2026, 08.00";
		for (const value of ["2026-10-24T15:00:00+07:00", "2026-10-24T08:00:00.123456789Z", new Date("2026-10-24T08:00:00Z"), 1792828800000]) {
			expect(formatDate(value, { timeZone: "UTC" })).toBe(expected);
		}
		expect(formatDate(0, { timeZone: "UTC" })).toBe("1 Jan 1970, 00.00");
	});
	test("does not silently interpret timezone-less strings or invalid calendar dates", () => {
		for (const value of [
			null,
			undefined,
			"",
			"bad",
			"2026-10-24",
			"2026-10-24T08:00:00",
			"2026-02-30T08:00:00Z",
			new Date(NaN),
			NaN,
			Infinity,
		]) {
			expect(toDate(value)).toBeNull();
			expect(formatDate(value)).toBe("—");
		}
		expect(formatDate(null, { fallback: "Tidak tersedia" })).toBe("Tidak tersedia");
	});
	test("does not mutate supplied Date instances", () => {
		const date = new Date("2026-10-24T08:00:00Z");
		const parsed = toDate(date);
		expect(parsed?.toISOString()).toBe("2026-10-24T08:00:00.000Z");
		parsed?.setFullYear(2000);
		expect(date.toISOString()).toBe("2026-10-24T08:00:00.000Z");
	});
	test("surfaces invalid formatter configuration rather than hiding programmer errors", () => {
		expect(() => formatDate("2026-10-24T08:00:00Z", { timeZone: "invalid/zone" })).toThrow(RangeError);
	});
});
