import { expect, test } from "bun:test";
import { createMemoryHistory, createRootRoute, createRoute, createRouter } from "@tanstack/react-router";
import { captureLinkToken } from "./token";
import { parseAuthSearch } from "../../router";

// a distinct query string gives a fresh module instance, i.e. a new document
const freshToken = (name: string) => import(`./token.ts?${name}`) as Promise<typeof import("./token")>;

function fakeWindow(url: string, state: unknown = { key: "fixture" }) {
	const location = new URL(url);
	const replaced: Array<[unknown, string, string]> = [];
	const win = {
		location: { search: location.search, pathname: location.pathname },
		history: { state, replaceState: (...args: [unknown, string, string]) => void replaced.push(args) },
	} as unknown as Pick<Window, "location" | "history">;
	return { win, replaced };
}

test("captures token and scrubs URL preserving history.state", () => {
	const { win, replaced } = fakeWindow("https://auth.invalid/verify-email?token=FAKE");
	expect(captureLinkToken(win)).toBe("FAKE");
	expect(replaced).toEqual([[{ key: "fixture" }, "", "/verify-email"]]);
});

test("second call in same document returns cached token", () => {
	const { win, replaced } = fakeWindow("https://auth.invalid/set-password?token=FAKE2");
	expect(captureLinkToken(win)).toBe("FAKE2");
	// the scrubbed URL no longer carries the token, as after a real replaceState
	const scrubbed = fakeWindow("https://auth.invalid/set-password");
	expect(captureLinkToken(scrubbed.win)).toBe("FAKE2");
	expect(replaced.length).toBe(1);
	expect(scrubbed.replaced.length).toBe(0);
});

test("capture returns null after scrub in a new document", async () => {
	const fresh = await freshToken("fresh-document");
	expect(fresh.captureLinkToken(fakeWindow("https://auth.invalid/reset-password").win)).toBeNull();
});

test("absent token returns null without replaceState", async () => {
	const fresh = await freshToken("absent-token");
	const { win, replaced } = fakeWindow("https://auth.invalid/verify-email?other=1");
	expect(fresh.captureLinkToken(win)).toBeNull();
	expect(replaced.length).toBe(0);
});

test("router state excludes token", async () => {
	const root = createRootRoute();
	const verify = createRoute({ getParentRoute: () => root, path: "/verify-email" });
	const router = createRouter({
		routeTree: root.addChildren([verify]),
		parseSearch: parseAuthSearch,
		history: createMemoryHistory({ initialEntries: ["/verify-email?token=FAKE&keep=1"] }),
	});
	await router.load();
	expect(router.state.location.search).toEqual({ keep: 1 });
	expect(router.state.matches.at(-1)?.search).toEqual({ keep: 1 });
});
