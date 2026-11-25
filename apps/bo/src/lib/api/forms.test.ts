import { expect, test } from "bun:test";
import { passwordValid } from "./forms";

test("password adapter counts Unicode codepoints with independent UTF-8 bound", () => {
	expect(passwordValid("a".repeat(14))).toBe(false);
	expect(passwordValid("a".repeat(15))).toBe(true);
	expect(passwordValid("😀".repeat(128))).toBe(true);
	expect(passwordValid("😀".repeat(129))).toBe(false);
	expect(passwordValid("a".repeat(129))).toBe(false);
});
