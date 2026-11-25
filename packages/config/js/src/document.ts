import { z } from "zod";
import { configError, type Diagnostic } from "./errors";
import { validateTemplates } from "./interpolation";

const scalar = z.union([z.string(), z.number().finite(), z.boolean(), z.null()]);
const basicAuth = z.strictObject({ enabled: scalar, username: scalar, password: scalar });
const database = z.strictObject({
	dsn: scalar,
	maxOpenConnections: scalar,
	connectionMaxLifetimeSeconds: scalar,
	connectionMaxIdleTimeSeconds: scalar,
	connectTimeoutSeconds: scalar,
});
const worker = z.strictObject({ shutdownTimeoutSeconds: scalar, maxWorkers: scalar });
const storage = z.strictObject({ endpoint: scalar, region: scalar, bucket: scalar, accessKeyId: scalar, secretAccessKey: scalar, presignTtlSeconds: scalar });
const service = {
	host: scalar,
	port: scalar,
	readTimeoutSeconds: scalar,
	writeTimeoutSeconds: scalar,
	idleTimeoutSeconds: scalar,
	shutdownTimeoutSeconds: scalar,
	docs: z.strictObject({ basicAuth }),
};
export const rawConfigSchema = z.strictObject({
	db: database,
	worker,
	storage,
	encryption: z.strictObject({ keys: z.array(scalar).min(1) }),
	app: z.strictObject({ environment: scalar, logLevel: scalar }),
	web: z.strictObject({ host: scalar, port: scalar, apiBaseUrl: scalar }),
	bo: z.strictObject({ host: scalar, port: scalar, boApiBaseUrl: scalar }),
	boApi: z.strictObject({
		...service,
		boOrigin: scalar,
		apiOrigin: scalar,
		development: scalar,
		trustedProxyCIDRs: z.array(scalar).max(64),
		csrfSecret: scalar,
	}),
	api: z.strictObject(service),
	webApi: z.strictObject({
		...service,
		identity: z.strictObject({
			authOrigin: scalar,
			development: scalar,
			trustedProxyCIDRs: z.array(scalar).max(64),
			csrfKey: scalar,
			limiterKey: scalar,
			email: z.strictObject({ from: scalar, resendKey: scalar }),
		}),
	}),
});
export type RawConfigDocument = z.output<typeof rawConfigSchema>;
export function validateConfigDocument(
	value: unknown,
	diagnostic: Diagnostic = { filePath: "input", runtime: "structure" },
): RawConfigDocument {
	const result = rawConfigSchema.safeParse(value);
	if (!result.success) throw configError(diagnostic, result.error.issues[0]?.path ?? [], "unknown key or invalid document shape");
	validateTemplates(result.data, diagnostic);
	return result.data;
}
