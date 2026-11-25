let captured: { path: string; token: string | null } | undefined;

/**
 * Reads the one-time link token from the initial URL and removes it from the address bar.
 * The token stays in module memory only so StrictMode's second effect still sees it.
 */
export function captureLinkToken(win: Pick<Window, "location" | "history">): string | null {
	const path = win.location.pathname;
	if (captured?.path === path) return captured.token;
	const token = new URLSearchParams(win.location.search).get("token") || null;
	if (token) win.history.replaceState(win.history.state, "", path);
	captured = { path, token };
	return token;
}
