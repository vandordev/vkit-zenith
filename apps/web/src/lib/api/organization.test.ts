import { expect, test } from "bun:test";
import type { AuthTransport } from "./client";
import { createAuthTransport } from "./client";
import { AuthError } from "./errors";
import {
	allPages,
	ListTooLarge,
	keepSameScope,
	listErrorKind,
	PagingStalled,
	orgErrorMessage,
	orgFailureMessage,
	orgQueryKey,
	orgWrite,
} from "./organization";

const config = { apiBaseUrl: "https://core.invalid" };
const fetcher = Object.assign(async () => Response.json({ token: "fixture-csrf" }), { preconnect: fetch.preconnect });
const transport = () => {
	const t = createAuthTransport(config, () => {}, fetcher);
	t.setCsrf("fixture-csrf");
	return t;
};

test("query keys carry identity epoch and organization", () => {
	const t = transport();
	const key = orgQueryKey(t, "org-1", "profile");
	expect(key).toEqual(["organization", 0, "org-1", "profile"]);
	expect(orgQueryKey(t, "org-2", "profile")).not.toEqual(key);
	t.identityChanged();
	expect(orgQueryKey(t, "org-1", "profile")).not.toEqual(key);
});

test("write carries the CSRF header and returns the result", async () => {
	const t = transport();
	expect(await orgWrite(t, async (headers) => headers["X-CSRF-Token"])).toBe("fixture-csrf");
});

test("write rejects when the identity changes while the CSRF token is awaited", async () => {
	let release: (response: Response) => void = () => {};
	const slow = Object.assign(() => new Promise<Response>((resolve) => (release = resolve)), { preconnect: fetch.preconnect });
	const t = createAuthTransport(config, () => {}, slow);
	let called = false;
	const write = orgWrite(t, async () => {
		called = true;
	});
	t.identityChanged();
	release(Response.json({ token: "stale" }));
	await expect(write).rejects.toMatchObject({ code: "session_changed" });
	expect(called).toBe(false);
});

test("a late write response after an identity change is dropped", async () => {
	const t = transport();
	const write = orgWrite(t, async () => {
		t.identityChanged();
		return "previous identity result";
	});
	await expect(write).rejects.toMatchObject({ code: "session_changed" });
});

test("error copy is fixed per status and never server detail", () => {
	const statuses = [401, 403, 404, 409, 422, 503, 0];
	const copy = statuses.map(orgErrorMessage);
	expect(new Set(copy).size).toBe(statuses.length);
	const hostile = new AuthError(422, "invalid", "SERVER-DETAIL-SENTINEL password=hunter2");
	expect(orgFailureMessage(hostile)).toBe(orgErrorMessage(422));
	expect(orgFailureMessage(hostile)).not.toContain("SENTINEL");
	expect(orgFailureMessage(new Error("SERVER-DETAIL-SENTINEL"))).toBe(orgErrorMessage(0));
});

const pages =
	(total: number, size = 100) =>
	async (page: number) => ({
		items: Array.from({ length: Math.max(0, Math.min(size, total - (page - 1) * size)) }, (_, index) => (page - 1) * size + index),
		pagination: { page, size, total: String(total) },
	});

test("allPages reads every page to the reported total, past 50 pages", async () => {
	expect(await allPages(pages(0))).toEqual([]);
	expect((await allPages(pages(250))).length).toBe(250);
	const big = await allPages(pages(5001));
	expect(big.length).toBe(5001);
	expect(big.at(-1)).toBe(5000);
});

test("allPages fails closed: the capacity limit is permanent, a stalled page sequence is a recoverable failure", async () => {
	// Total beyond the page guard: capacity, no retry.
	const tooLarge = await allPages(pages(30000)).catch((error) => error);
	expect(tooLarge).toBeInstanceOf(ListTooLarge);
	expect(listErrorKind(tooLarge)).toBe("tooLarge");
	// An empty page before the reported total (101 records, second page empty): inconsistent, not "too many", and retryable.
	const short = async (page: number) => ({
		items: page === 1 ? Array.from({ length: 100 }, (_, i) => i) : [],
		pagination: { page, size: 100, total: "101" },
	});
	const stalled = await allPages(short).catch((error) => error);
	expect(stalled).toBeInstanceOf(PagingStalled);
	expect(stalled).not.toBeInstanceOf(ListTooLarge);
	expect(listErrorKind(stalled)).toBe("failed");
	// The same list succeeds once the server is consistent again (the retry).
	expect((await allPages(pages(101))).length).toBe(101);
	// A failing page propagates as an ordinary failure.
	const failure = await allPages(async () => Promise.reject(new Error("503"))).catch((error) => error);
	expect(listErrorKind(failure)).toBe("failed");
});

test("keepSameScope keeps the previous page only inside the same epoch and organization", () => {
	const t = { epoch: () => 3 } as unknown as AuthTransport;
	const key = (epoch: number, org: string) => ({ queryKey: ["organization", epoch, org, "api-keys", 1] });
	const keep = keepSameScope(t, "org-a");
	expect(keep("rows", key(3, "org-a"))).toBe("rows");
	expect(keep("rows", key(3, "org-b"))).toBeUndefined();
	expect(keep("rows", key(2, "org-a"))).toBeUndefined();
	expect(keep("rows", undefined)).toBeUndefined();
});
