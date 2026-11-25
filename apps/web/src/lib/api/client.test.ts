import { afterEach, expect, setSystemTime, test } from "bun:test";
import { authV1GetCsrf, authV1Login } from "./generated/sdk.gen";
import { createAuthTransport } from "./client";
import { AuthError, toFormStatus } from "./errors";

const config = { apiBaseUrl: "https://core.invalid" };
const body = { email: "fixture@example.invalid", password: "fixture-password" };
afterEach(() => setSystemTime());

function fake(handler: (request: Request) => Response | Promise<Response>) {
	const calls: Request[] = [];
	const fetcher = Object.assign(
		async (input: Parameters<typeof fetch>[0], init?: Parameters<typeof fetch>[1]) => {
			const request = new Request(input, init);
			calls.push(request);
			return handler(request);
		},
		{ preconnect: fetch.preconnect },
	);
	return { calls, fetcher };
}
const csrfCalls = (calls: Request[]) => calls.filter((request) => request.url.endsWith("/v1/auth/csrf"));
const login = async (t: ReturnType<typeof createAuthTransport>) =>
	authV1Login({ client: t.client, throwOnError: true, headers: { "X-CSRF-Token": await t.csrf() }, body });
let minted = 0;
const mint = (request: Request) =>
	request.url.endsWith("/v1/auth/csrf") ? Response.json({ token: `fixture-${++minted}` }) : Response.json({ kind: "full" });

test("bootstrap once for concurrent unsafe calls", async () => {
	const { calls, fetcher } = fake(mint);
	const t = createAuthTransport(config, () => {}, fetcher);
	await Promise.all([login(t), login(t), login(t)]);
	expect(csrfCalls(calls).length).toBe(1);
});

test("setCsrf replaces token without fetch", async () => {
	const { calls, fetcher } = fake(mint);
	const t = createAuthTransport(config, () => {}, fetcher);
	t.setCsrf("fixture-set");
	expect(await t.csrf()).toBe("fixture-set");
	expect(calls.length).toBe(0);
});

test("csrf refetched after 9 minutes", async () => {
	const { calls, fetcher } = fake(mint);
	const t = createAuthTransport(config, () => {}, fetcher);
	setSystemTime(new Date("2026-01-01T00:00:00Z"));
	await t.csrf();
	setSystemTime(new Date("2026-01-01T00:08:59Z"));
	await t.csrf();
	expect(csrfCalls(calls).length).toBe(1);
	setSystemTime(new Date("2026-01-01T00:09:00Z"));
	await t.csrf();
	expect(csrfCalls(calls).length).toBe(2);
	t.setCsrf("fixture-set");
	setSystemTime(new Date("2026-01-01T00:17:59Z"));
	expect(await t.csrf()).toBe("fixture-set");
});

test("403 clears token and does not retry", async () => {
	const { calls, fetcher } = fake((request) =>
		request.url.endsWith("/csrf") ? mint(request) : Response.json({ title: "Forbidden", detail: "csrf" }, { status: 403 }),
	);
	const t = createAuthTransport(config, () => {}, fetcher);
	await expect(login(t)).rejects.toBeInstanceOf(AuthError);
	expect(calls.length).toBe(2);
	await t.csrf();
	expect(csrfCalls(calls).length).toBe(2);
});

test("no unsafe replay after 403", async () => {
	let unsafe = 0;
	const { fetcher } = fake((request) => {
		if (request.url.endsWith("/csrf")) return mint(request);
		unsafe++;
		return Response.json({ title: "Forbidden" }, { status: 403 });
	});
	const t = createAuthTransport(config, () => {}, fetcher);
	await expect(login(t)).rejects.toBeInstanceOf(AuthError);
	expect(unsafe).toBe(1);
});

test("bootstrap failure surfaces unavailable", async () => {
	const { fetcher } = fake(() => Response.json({ title: "Service Unavailable" }, { status: 503 }));
	const t = createAuthTransport(config, () => {}, fetcher);
	const error = await t.csrf().catch((e: unknown) => e);
	expect(toFormStatus(error)).toEqual({ kind: "unavailable" });
	const down = createAuthTransport(
		config,
		() => {},
		Object.assign(() => Promise.reject(new TypeError("fetch failed")), { preconnect: fetch.preconnect }),
	);
	expect(toFormStatus(await down.csrf().catch((e: unknown) => e))).toEqual({ kind: "unavailable" });
});

test("epoch change mid-flight rejects stale response", async () => {
	let finish: ((response: Response) => void) | undefined;
	const fetcher = Object.assign(() => new Promise<Response>((resolve) => (finish = resolve)), { preconnect: fetch.preconnect });
	let changed = 0;
	const t = createAuthTransport(config, () => changed++, fetcher);
	const before = t.epoch();
	const pending = authV1GetCsrf({ client: t.client, throwOnError: true });
	await Promise.resolve();
	await Promise.resolve();
	t.identityChanged();
	finish?.(Response.json({ token: "stale" }));
	const error = await pending.catch((e: unknown) => e);
	expect(error).toBeInstanceOf(AuthError);
	expect((error as AuthError).code).toBe("session_changed");
	expect(t.epoch()).toBe(before + 1);
	expect(changed).toBe(1);
});

test("422 maps field errors", async () => {
	const { fetcher } = fake((request) =>
		request.url.endsWith("/csrf")
			? mint(request)
			: Response.json(
					{ title: "Unprocessable Entity", status: 422, errors: [{ message: "bukan email", location: "body.email" }] },
					{ status: 422 },
				),
	);
	const t = createAuthTransport(config, () => {}, fetcher);
	const status = toFormStatus(await login(t).catch((e: unknown) => e));
	expect(status.kind).toBe("error");
	expect(status.kind === "error" && status.fields).toEqual({ email: "bukan email" });
});

test("429 maps retryAfter", async () => {
	const { fetcher } = fake((request) =>
		request.url.endsWith("/csrf") ? mint(request) : Response.json({ title: "Too Many" }, { status: 429, headers: { "Retry-After": "30" } }),
	);
	const t = createAuthTransport(config, () => {}, fetcher);
	expect(toFormStatus(await login(t).catch((e: unknown) => e))).toEqual({ kind: "rate-limited", retryAfterSeconds: 30 });
});

test("requests use credentials include and no-store", async () => {
	const { calls, fetcher } = fake(mint);
	const t = createAuthTransport(config, () => {}, fetcher);
	await login(t);
	expect(calls.length).toBe(2);
	expect(calls.every((request) => request.credentials === "include" && request.cache === "no-store")).toBe(true);
});

// Headers arrive in the originating epoch; the body is released only after the identity changed.
function streamed(status: number, json: unknown) {
	let release = () => {};
	const stream = new ReadableStream<Uint8Array>({
		start(controller) {
			release = () => {
				controller.enqueue(new TextEncoder().encode(JSON.stringify(json)));
				controller.close();
			};
		},
	});
	return { release, response: new Response(stream, { status, headers: { "Content-Type": "application/json" } }) };
}
const tick = () => new Promise((resolve) => setTimeout(resolve, 0));

test("stale body after identity change rejects session_changed on direct SDK call", async () => {
	const held = streamed(200, { token: "stale" });
	const { calls, fetcher } = fake(() => held.response);
	const t = createAuthTransport(config, () => {}, fetcher);
	const pending = authV1GetCsrf({ client: t.client, throwOnError: true });
	await tick();
	t.identityChanged();
	t.setCsrf("fixture-new");
	held.release();
	const error = await pending.catch((e: unknown) => e);
	expect(error).toBeInstanceOf(AuthError);
	expect((error as AuthError).code).toBe("session_changed");
	expect(await t.csrf()).toBe("fixture-new");
	expect(calls.length).toBe(1);
});

test("stale csrf bootstrap rejects without overwriting the new token or the new acquisition", async () => {
	const stale = streamed(200, { token: "stale" });
	const fresh = streamed(200, { token: "fixture-fresh" });
	const queue = [stale, fresh];
	const { calls, fetcher } = fake(() => queue.shift()!.response);
	const t = createAuthTransport(config, () => {}, fetcher);
	const old = t.csrf().catch((e: unknown) => e);
	await tick();
	t.identityChanged();
	const next = t.csrf();
	const shared = t.csrf();
	await tick();
	expect(calls.length).toBe(2);
	stale.release();
	const error = await old;
	expect((error as AuthError).code).toBe("session_changed");
	fresh.release();
	expect(await next).toBe("fixture-fresh");
	expect(await shared).toBe("fixture-fresh");
	expect(calls.length).toBe(2);
	expect(await t.csrf()).toBe("fixture-fresh");
});

test("stale 403 body after identity change rejects session_changed and keeps the new token", async () => {
	const held = streamed(403, { title: "Forbidden" });
	const { calls, fetcher } = fake(() => held.response);
	const t = createAuthTransport(config, () => {}, fetcher);
	const pending = authV1Login({ client: t.client, throwOnError: true, headers: { "X-CSRF-Token": "fixture-old" }, body }).catch(
		(e: unknown) => e,
	);
	await tick();
	t.identityChanged();
	t.setCsrf("fixture-new");
	held.release();
	expect(((await pending) as AuthError).code).toBe("session_changed");
	expect(calls.length).toBe(1);
	expect(await t.csrf()).toBe("fixture-new");
	expect(calls.length).toBe(1);
});
