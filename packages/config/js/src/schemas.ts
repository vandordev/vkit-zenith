import { z } from "zod";
import { apiBaseUrlSchema } from "./public";

export const appEnvironmentSchema = z.enum(["development", "staging", "production", "test"]);
export type AppEnvironment = z.output<typeof appEnvironmentSchema>;
const numeric = z
	.union([z.number(), z.string().regex(/^\d+$/)])
	.transform(Number)
	.pipe(z.number().int());
const port = numeric.pipe(z.number().min(1).max(65535));
const seconds = numeric.pipe(z.number().min(1).max(3600));
export const docsEnabledSchema = z
	.union([z.boolean(), z.literal("true"), z.literal("false")])
	.transform((value) => value === true || value === "true");
const app = z.strictObject({ environment: appEnvironmentSchema, logLevel: z.enum(["debug", "info", "warn", "error"]) });
const listener = { host: z.string().min(1), port };
const basicAuth = z
	.strictObject({ enabled: docsEnabledSchema, username: z.string(), password: z.string() })
	.refine((value) => !value.enabled || (value.username.length > 0 && value.password.length > 0), {
		message: "enabled documentation protection requires credentials",
	});
const service = {
	...listener,
	readTimeoutSeconds: seconds,
	writeTimeoutSeconds: seconds,
	idleTimeoutSeconds: seconds,
	shutdownTimeoutSeconds: seconds,
	docs: z.strictObject({ basicAuth }),
};
export const webToolConfigSchema = z.strictObject({ app, web: z.strictObject(listener) });
export const boToolConfigSchema = z.strictObject({ app, bo: z.strictObject(listener) });
export const runtimeSchemas = {
	bo: z.strictObject({ app, bo: z.strictObject({ ...listener, boApiBaseUrl: apiBaseUrlSchema }) }),
	web: z.strictObject({ app, web: z.strictObject({ ...listener, apiBaseUrl: apiBaseUrlSchema }) }),
	api: z.strictObject({ app, api: z.strictObject(service) }),
} as const;
export type RuntimeName = keyof typeof runtimeSchemas;
export type RuntimeConfig<R extends RuntimeName> = z.output<(typeof runtimeSchemas)[R]>;
