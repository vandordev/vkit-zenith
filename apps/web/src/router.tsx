import { createRouter, defaultParseSearch } from "@tanstack/react-router";
import { AuthApiProvider } from "./lib/api/provider";
import { routeTree } from "./routeTree.gen";

// One-time link tokens must never reach Router state; the history entry keeps the raw URL until token capture scrubs it.
export const parseAuthSearch = (search: string) => {
	const value: Record<string, unknown> = defaultParseSearch(search);
	delete value.token;
	return value;
};

export function getRouter() {
	return createRouter({
		routeTree,
		parseSearch: parseAuthSearch,
		Wrap: AuthApiProvider,
		defaultPreload: "intent",
		scrollRestoration: true,
		// The inset dashboard panel scrolls on desktop instead of the window (AppShell `inset`).
		scrollToTopSelectors: ['[data-scroll-restoration-id="app-shell-main"]'],
	});
}
declare module "@tanstack/react-router" {
	interface Register {
		router: ReturnType<typeof getRouter>;
	}
}
