import { expect, test } from "bun:test";
import { amountProblem, toBasisPoints, toMinorUnits } from "./form-input";

test("minor units are exact for IDR and USD and never rounded", () => {
	expect(toMinorUnits("125000", "IDR")).toBe("125000");
	expect(toMinorUnits("0", "IDR")).toBe("0");
	expect(toMinorUnits("0012", "IDR")).toBe("12");
	expect(toMinorUnits("12.5", "USD")).toBe("1250");
	expect(toMinorUnits("0.07", "USD")).toBe("7");
	expect(toMinorUnits("12", "USD")).toBe("1200");
	expect(toMinorUnits("4.35", "USD")).toBe("435");
	expect(toMinorUnits("12.505", "USD")).toBeUndefined();
	expect(toMinorUnits("1.5", "IDR")).toBeUndefined();
	expect(toMinorUnits("1e3", "IDR")).toBeUndefined();
	expect(toMinorUnits("-1", "IDR")).toBeUndefined();
	expect(toMinorUnits("", "IDR")).toBeUndefined();
	expect(toMinorUnits("1", "XXXX")).toBeUndefined();
	expect(toMinorUnits("9999999999999999999", "IDR")).toBeUndefined();
	expect(toMinorUnits("999999999999999999", "IDR")).toBe("999999999999999999");
});

test("exponents zero, two and three: JPY, EUR/USD, KWD", () => {
	expect(toMinorUnits("1500", "JPY")).toBe("1500");
	expect(toMinorUnits("1.5", "JPY")).toBeUndefined();
	expect(toMinorUnits("19.99", "EUR")).toBe("1999");
	expect(toMinorUnits("19.999", "EUR")).toBeUndefined();
	expect(toMinorUnits("1.234", "KWD")).toBe("1234");
	expect(toMinorUnits("1", "KWD")).toBe("1000");
	expect(toMinorUnits("1.2345", "KWD")).toBeUndefined();
	expect(toMinorUnits("0", "KWD")).toBe("0");
	// 18 minor-unit digits is the contract's limit: a three-decimal amount overflows it three digits sooner than a zero-decimal one.
	expect(toMinorUnits("999999999999999", "KWD")).toBe("999999999999999000");
	expect(toMinorUnits("9999999999999999", "KWD")).toBeUndefined();
	expect(toMinorUnits("999999999999999.999", "KWD")).toBe("999999999999999999");
});

test("amount problems name the currency's precision", () => {
	expect(amountProblem("JPY")).toContain("bulat");
	expect(amountProblem("USD")).toContain("2 desimal");
	expect(amountProblem("KWD")).toContain("3 desimal");
});

test("percent converts to basis points within 1–10000", () => {
	expect(toBasisPoints("10")).toBe(1000);
	expect(toBasisPoints("12.5")).toBe(1250);
	expect(toBasisPoints("0.01")).toBe(1);
	expect(toBasisPoints("100")).toBe(10000);
	expect(toBasisPoints("100.01")).toBeUndefined();
	expect(toBasisPoints("0")).toBeUndefined();
	expect(toBasisPoints("1.005")).toBeUndefined();
	expect(toBasisPoints("abc")).toBeUndefined();
});
