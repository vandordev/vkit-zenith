import { createServerFn } from "@tanstack/react-start";
export const getPublicConfig = createServerFn({ method: "GET" }).handler(async () => {
	const { readPublicConfig } = await import("./public-config.server");
	return readPublicConfig();
});
