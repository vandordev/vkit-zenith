import { expect, test } from "bun:test";
import { singleFlight } from "./single-flight";

test("singleFlight ignores calls while pending", async () => {
	let release: () => void = () => {};
	let calls = 0;
	const run = singleFlight(async () => {
		calls++;
		await new Promise<void>((resolve) => (release = resolve));
	});
	run();
	run();
	run();
	expect(calls).toBe(1);
	release();
	await Promise.resolve();
});

test("allows next call after settle", async () => {
	let calls = 0;
	const run = singleFlight(async () => {
		calls++;
	});
	run();
	await new Promise((resolve) => setTimeout(resolve, 0));
	run();
	expect(calls).toBe(2);
});
