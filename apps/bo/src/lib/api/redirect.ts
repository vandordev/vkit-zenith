// Relative internal destinations only. Never interpret a supplied host/origin.
export function safeReturnTo(value: unknown): string {
	if (typeof value !== "string" || !value.startsWith("/") || value.startsWith("//") || /[\\\u0000-\u0020\u007f]/u.test(value)) return "/";
	try {
		const decoded = decodeURIComponent(value);
		if (decoded.startsWith("//") || /[\\\u0000-\u0020\u007f]/u.test(decoded)) return "/";
		const url = new URL(value, "https://bo.invalid");
		if (url.origin !== "https://bo.invalid" || /^\/login(?:\/|$)/i.test(decodeURIComponent(url.pathname))) return "/";
		return `${url.pathname}${url.search}${url.hash}`;
	} catch {
		return "/";
	}
}
