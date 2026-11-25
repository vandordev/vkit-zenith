import { createServerFn } from "@tanstack/react-start";
export const getBOPublicConfig = createServerFn({ method: "GET" }).handler(async () => {
	const { readBOPublicConfig } = await import("./public-config.server");
	return readBOPublicConfig();
});
