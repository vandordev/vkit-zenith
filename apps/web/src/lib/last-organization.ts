const KEY = "zenith-last-organization";

// Pure per-browser convenience: storage may be blocked, so every access is guarded.
export function readLastOrganization(): string | undefined {
	try {
		return localStorage.getItem(KEY) ?? undefined;
	} catch {
		return undefined;
	}
}

export function writeLastOrganization(slug: string): void {
	try {
		localStorage.setItem(KEY, slug);
	} catch {
		// shortcut: ignored when storage is blocked; the entry falls back to the first organization
	}
}
