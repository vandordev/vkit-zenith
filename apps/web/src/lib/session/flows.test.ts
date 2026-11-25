import { expect, test } from "bun:test";
import { createAuthTransport } from "../api/client";
import { checkEmailState, confirm, inspectForRoute, login, logout, register, resolveHome } from "./flows";

const config = { apiBaseUrl: "https://core.invalid" };
const session = {
	kind: "full",
	displayName: "Fixture",
	email: "fixture@example.invalid",
	emailVerified: true,
	authenticatedAt: "x",
	expiresAt: "x",
};
const problem = (status: number) => Response.json({ title: "fixture", status }, { status });

type Handler = (request: Request) => Response | Promise<Response>;
/** Routes by "METHOD /path"; unmatched requests fail the test loudly. */
function setup(routes: Record<string, Handler>) {
	const calls: Request[] = [];
	const bodies: unknown[] = [];
	const fetcher = Object.assign(
		async (input: Parameters<typeof fetch>[0], init?: Parameters<typeof fetch>[1]) => {
			const request = new Request(input, init);
			calls.push(request);
			if (request.method !== "GET")
				bodies.push(
					await request
						.clone()
						.json()
						.catch(() => undefined),
				);
			const key = `${request.method} ${new URL(request.url).pathname}`;
			if (key === "GET /v1/auth/csrf" && !routes[key]) return Response.json({ token: "fixture-csrf" });
			const handler = routes[key];
			if (!handler) throw new Error(`unexpected ${key}`);
			return handler(request);
		},
		{ preconnect: fetch.preconnect },
	);
	const t = createAuthTransport(config, () => {}, fetcher);
	const count = (key: string) => calls.filter((r) => `${r.method} ${new URL(r.url).pathname}` === key).length;
	return { t, calls, bodies, count };
}
const ok = (body: unknown) => () => Response.json(body);
const touchOk = ok({ expiresAt: "x" });

test("home full touches once", async () => {
	const { t, count } = setup({ "GET /v1/auth/session/current": ok(session), "POST /v1/auth/session/touch": touchOk });
	expect(await resolveHome(t)).toEqual({ kind: "signed-in", displayName: "Fixture", email: "fixture@example.invalid" });
	expect(count("POST /v1/auth/session/touch")).toBe(1);
});

test("home touch 429 still signed-in", async () => {
	const { t } = setup({ "GET /v1/auth/session/current": ok(session), "POST /v1/auth/session/touch": () => problem(429) });
	expect((await resolveHome(t)).kind).toBe("signed-in");
});

test("home restricted", async () => {
	const { t, count } = setup({ "GET /v1/auth/session/current": ok({ ...session, kind: "restricted" }) });
	expect((await resolveHome(t)).kind).toBe("restricted");
	expect(count("POST /v1/auth/session/touch")).toBe(0);
});

test("login restricted sets csrf and changes epoch", async () => {
	const { t, count } = setup({ "POST /v1/auth/login": ok({ ...session, kind: "restricted", csrfToken: "fixture-login" }) });
	expect(await login(t, { email: "fixture@example.invalid", password: "fixture-password" })).toEqual({ kind: "restricted" });
	expect(t.epoch()).toBe(1);
	expect(await t.csrf()).toBe("fixture-login");
	expect(count("GET /v1/auth/csrf")).toBe(1);
});

test("register sends +E164 and changes epoch on accepted", async () => {
	const { t, bodies } = setup({ "POST /v1/auth/register": () => Response.json({ accepted: true }, { status: 202 }) });
	const values = {
		displayName: "Fixture",
		email: "fixture@example.invalid",
		telephoneDigits: "6281234567890",
		password: "fixture-password",
	};
	expect(await register(t, values)).toEqual({ kind: "accepted" });
	expect(bodies.at(-1)).toMatchObject({ telephone: "+6281234567890" });
	expect(t.epoch()).toBe(1);
});

test("setup confirm sends +E164", async () => {
	const { t, bodies } = setup({ "POST /v1/auth/action/confirm": ok({ completed: true, requiresFreshLogin: true }) });
	await confirm(t, "FAKE", { password: "fixture-password", telephoneDigits: "6281234567890" });
	expect(bodies.at(-1)).toEqual({ token: "FAKE", password: "fixture-password", telephone: "+6281234567890" });
});

test("checkEmailState 401 → generic", async () => {
	const { t } = setup({ "GET /v1/auth/session/current": () => problem(401) });
	expect(await checkEmailState(t)).toEqual({ kind: "generic" });
});

test("logout failure keeps session and epoch", async () => {
	const { t } = setup({ "POST /v1/auth/session/logout": () => problem(503) });
	expect(await logout(t)).toEqual({ kind: "unavailable" });
	expect(t.epoch()).toBe(0);
});

const inspected = (body: object) => ({
	"GET /v1/auth/action/inspect": ok({ expiresAt: "2026-01-01T00:00:00Z", requiresLogin: false, ...body }),
});

test("inspect purpose mismatch → unavailable and no confirm", async () => {
	const { t, count } = setup(inspected({ purpose: "reset_password" }));
	expect(await inspectForRoute(t, "FAKE", "verify_email")).toEqual({ kind: "outcome", outcome: "unavailable" });
	expect(count("POST /v1/auth/action/confirm")).toBe(0);
});

test("inspect requiresLogin → cross-browser", async () => {
	const { t } = setup(inspected({ purpose: "verify_email", requiresLogin: true }));
	expect(await inspectForRoute(t, "FAKE", "verify_email")).toEqual({ kind: "cross-browser" });
});

test("confirm verify with csrf → completed-signed-in", async () => {
	const { t } = setup({
		"POST /v1/auth/action/confirm": ok({ completed: true, requiresFreshLogin: false, csrfToken: "fixture-verified" }),
	});
	expect(await confirm(t, "FAKE")).toEqual({ kind: "outcome", outcome: "completed-signed-in" });
	expect(t.epoch()).toBe(1);
	expect(await t.csrf()).toBe("fixture-verified");
});

test("confirm reset → completed-login", async () => {
	const { t } = setup({ "POST /v1/auth/action/confirm": ok({ completed: true, requiresFreshLogin: true }) });
	expect(await confirm(t, "FAKE", { password: "fixture-password" })).toEqual({ kind: "outcome", outcome: "completed-login" });
});

// Deferred touch: identity changes while the touch request is in flight.
function deferredTouch(extra: Record<string, Handler> = {}) {
	let release: () => void = () => {};
	const gate = new Promise<void>((resolve) => (release = resolve));
	const s = setup({
		"GET /v1/auth/session/current": ok(session),
		"POST /v1/auth/session/touch": async () => {
			await gate;
			return Response.json({ expiresAt: "x" });
		},
		...extra,
	});
	return { ...s, release };
}
const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

test("home identity change during touch never returns the old identity", async () => {
	const { t, release } = deferredTouch();
	const pending = resolveHome(t);
	await settle();
	t.identityChanged();
	release();
	expect(await pending).toEqual({ kind: "unavailable" });
});

test("home identity change between steps never returns the old identity", async () => {
	const { t, release } = deferredTouch();
	release();
	const csrf = t.csrf;
	t.csrf = async () => {
		const token = await csrf();
		t.identityChanged();
		return token;
	};
	expect((await resolveHome(t)).kind).toBe("unavailable");
});
