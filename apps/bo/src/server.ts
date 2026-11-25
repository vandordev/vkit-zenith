import handler, { createServerEntry } from "@tanstack/react-start/server-entry";
import { configureApplicationRoot } from "./server/bootstrap.server";
configureApplicationRoot();
export default createServerEntry({
	fetch: async (request) => {
		const response = await handler.fetch(request);
		response.headers.set("Cache-Control", "no-store");
		response.headers.set("X-Robots-Tag", "noindex, nofollow");
		response.headers.set("Referrer-Policy", "no-referrer");
		return response;
	},
});
