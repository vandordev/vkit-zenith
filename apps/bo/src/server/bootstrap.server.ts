import "@tanstack/react-start/server-only";
import { setApplicationRoot } from "@repo/config/server";
export function configureApplicationRoot() {
	if (process.env.ZENITH_APPLICATION_ROOT) setApplicationRoot(process.env.ZENITH_APPLICATION_ROOT);
	// Nitro copies unresolved config/config.yaml into the Vercel function root.
	else if (process.env.VERCEL === "1") setApplicationRoot(process.cwd());
}
