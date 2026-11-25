import handler, { createServerEntry } from "@tanstack/react-start/server-entry";
import { configureApplicationRoot } from "./server/bootstrap.server";

configureApplicationRoot();

export default createServerEntry({
	fetch: async (request) => {
		// Generate correlation locally; never log untrusted headers, URLs or errors.
		const requestId = crypto.randomUUID();
		const startedAt = Date.now();
		const response = await handler.fetch(request);
		console.info(JSON.stringify({ service: "web", requestId, durationMs: Date.now() - startedAt, status: response.status }));
		response.headers.set("x-request-id", requestId);
		response.headers.set("Cache-Control", "no-store");
		response.headers.set("Referrer-Policy", "no-referrer");
		return response;
	},
});
