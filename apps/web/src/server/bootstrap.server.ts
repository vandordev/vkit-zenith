import "@tanstack/react-start/server-only";
import { setApplicationRoot } from "@repo/config/server";

// Narrow deployment-root adapter: Vercel executes from the function root where
// Nitro's compiled hook copies config/config.yaml. Containers use launcher /app.
// This does not resolve configuration, discover env files, or expose env to UI.
export function configureApplicationRoot() {
	if (process.env.ZENITH_APPLICATION_ROOT) setApplicationRoot(process.env.ZENITH_APPLICATION_ROOT);
	else if (process.env.VERCEL === "1") setApplicationRoot(process.cwd());
}
