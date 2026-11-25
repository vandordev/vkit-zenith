import { expect, test } from "bun:test";
import { boV1GetMe, boV1Login } from "./generated/sdk.gen";
import { BOError, createBOTransport } from "./client";

test("401 login clears CSRF; next explicit attempt reacquires without replay", async () => {
	const calls: Request[] = [];
	let minted = 0;
	let lost = 0;
	const fetcher = Object.assign(
		async (input: Parameters<typeof fetch>[0]) => {
			const request = new Request(input);
			calls.push(request);
			if (request.url.endsWith("/csrf")) return Response.json({ token: `fixture-${++minted}` });
			return Response.json({ type: "urn:zenith:bo:unauthenticated", detail: "Email atau kata sandi salah." }, { status: 401 });
		},
		{ preconnect: fetch.preconnect },
	);
	const transport = createBOTransport({ boApiBaseUrl: "https://api.invalid" }, () => lost++, fetcher);
	for (let i = 0; i < 2; i++) {
		await expect(
			boV1Login({ ...(await transport.writeOptions()), body: { email: "fixture@example.invalid", password: "" } }),
		).rejects.toBeInstanceOf(BOError);
	}
	expect(calls.length).toBe(4);
	expect(minted).toBe(2);
	expect(lost).toBe(2);
	expect(calls.every((request) => request.credentials === "include")).toBe(true);
	expect(calls[1]?.headers.get("X-CSRF-Token")).toBe("fixture-1");
	expect(calls[3]?.headers.get("X-CSRF-Token")).toBe("fixture-2");
});

test("late protected response from prior identity is rejected after reset", async () => {
	let finish: ((response: Response) => void) | undefined;
	const fetcher = Object.assign(
		() =>
			new Promise<Response>((resolve) => {
				finish = resolve;
			}),
		{ preconnect: fetch.preconnect },
	);
	const transport = createBOTransport({ boApiBaseUrl: "https://api.invalid" }, () => {}, fetcher);
	const pending = boV1GetMe({ client: transport.client, throwOnError: true });
	await Promise.resolve();
	await Promise.resolve();
	transport.reset();
	if (!finish) throw new Error("request not started");
	finish(Response.json({ id: "prior-user" }));
	await expect(pending).rejects.toBeInstanceOf(BOError);
});

test("403 refreshes CSRF only on next explicit submission, never replays mutation", async () => {
	let minted = 0;
	let mutations = 0;
	const fetcher = Object.assign(
		async (input: Parameters<typeof fetch>[0]) => {
			const request = new Request(input);
			if (request.url.endsWith("/csrf")) return Response.json({ token: `fixture-${++minted}` });
			mutations++;
			return Response.json({ type: "urn:zenith:bo:denied", detail: "Akses ditolak." }, { status: 403 });
		},
		{ preconnect: fetch.preconnect },
	);
	const transport = createBOTransport({ boApiBaseUrl: "https://api.invalid" }, () => {}, fetcher);
	await expect(
		boV1Login({ ...(await transport.writeOptions()), body: { email: "fixture@example.invalid", password: "" } }),
	).rejects.toBeInstanceOf(BOError);
	expect(mutations).toBe(1);
	expect(minted).toBe(1);
	await transport.writeOptions();
	expect(minted).toBe(2);
	expect(mutations).toBe(1);
});
