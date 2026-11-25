import { createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
export function getRouter() {
	return createRouter({
		routeTree,
		defaultPreload: "intent",
		scrollRestoration: true,
		// The inset shell's panel scrolls on desktop instead of the window (AppShell `inset`).
		scrollToTopSelectors: ['[data-scroll-restoration-id="app-shell-main"]'],
	});
}
declare module "@tanstack/react-router" {
	interface Register {
		router: ReturnType<typeof getRouter>;
	}
}
