import type { z } from "zod";
import type { webToolConfigSchema, boToolConfigSchema } from "./schemas";
export function webToolEnvironment(config: z.output<typeof webToolConfigSchema>) {
	return { PORT: String(config.web.port), HOSTNAME: config.web.host, NODE_ENV: config.app.environment };
}
export function boToolEnvironment(config: z.output<typeof boToolConfigSchema>) {
	return { PORT: String(config.bo.port), HOSTNAME: config.bo.host, NODE_ENV: config.app.environment };
}
