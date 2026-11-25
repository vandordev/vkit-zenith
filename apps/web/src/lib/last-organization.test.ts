import { afterEach, expect, test } from "bun:test";
import { readLastOrganization, writeLastOrganization } from "./last-organization";

const original = Object.getOwnPropertyDescriptor(globalThis, "localStorage");
afterEach(() => {
	if (original) Object.defineProperty(globalThis, "localStorage", original);
	else Reflect.deleteProperty(globalThis, "localStorage");
});
const store = (value: unknown) => Object.defineProperty(globalThis, "localStorage", { value, configurable: true });

test("remembers the last organization when storage works", () => {
	const data = new Map<string, string>();
	store({ getItem: (key: string) => data.get(key) ?? null, setItem: (key: string, value: string) => data.set(key, value) });
	expect(readLastOrganization()).toBeUndefined();
	writeLastOrganization("kopi-senja");
	expect(readLastOrganization()).toBe("kopi-senja");
});

test("blocked storage falls back without throwing", () => {
	Object.defineProperty(globalThis, "localStorage", {
		configurable: true,
		get() {
			throw new DOMException("blocked", "SecurityError");
		},
	});
	expect(readLastOrganization()).toBeUndefined();
	expect(() => writeLastOrganization("kopi-senja")).not.toThrow();
});
