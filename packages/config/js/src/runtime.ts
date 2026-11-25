import { z } from "zod";
import { loadConfigDocument } from "./loader";
import type { RawConfigDocument } from "./document";
import { interpolateSelected } from "./interpolation";
import { configError } from "./errors";
import { configFilePath } from "./paths";
import { publicBOConfigSchema, publicWebConfigSchema } from "./public";
import {
	docsEnabledSchema,
	runtimeSchemas,
	webToolConfigSchema,
	boToolConfigSchema,
	type RuntimeName,
	type RuntimeConfig,
} from "./schemas";

export type ConfigOptions = { filePath?: string; environment?: Readonly<Record<string, string | undefined>> };
type Snapshot = { raw: RawConfigDocument; environment: Readonly<Record<string, string | undefined>>; filePath: string };
function snapshot(options: ConfigOptions, runtime: string): Snapshot {
	const filePath = options.filePath ?? configFilePath();
	return { raw: loadConfigDocument(filePath, runtime), environment: { ...(options.environment ?? process.env) }, filePath };
}
function parse<S extends z.ZodType>(schema: S, selected: unknown, input: Snapshot, runtime: string): z.output<S> {
	const diagnostic = { filePath: input.filePath, runtime };
	const result = schema.safeParse(interpolateSelected(selected, input.environment, diagnostic));
	if (!result.success) throw configError(diagnostic, result.error.issues[0]?.path ?? [], "invalid value or incomplete credentials");
	return result.data;
}
function selectRuntime(runtime: RuntimeName, input: Snapshot) {
	const raw = input.raw;
	if (runtime === "web") return { app: raw.app, web: raw.web };
	if (runtime === "bo") return { app: raw.app, bo: raw.bo };
	const service = raw[runtime];
	// Resolve the switch first. Disabled credentials are structurally checked
	// above, but their environment templates/values are never selected.
	const enabled = parse(docsEnabledSchema, service.docs.basicAuth.enabled, input, runtime);
	const basicAuth = enabled ? { ...service.docs.basicAuth, enabled } : { enabled, username: "", password: "" };
	// New identity sections remain opaque to the historical selectors.
	const selected = {
		host: service.host,
		port: service.port,
		readTimeoutSeconds: service.readTimeoutSeconds,
		writeTimeoutSeconds: service.writeTimeoutSeconds,
		idleTimeoutSeconds: service.idleTimeoutSeconds,
		shutdownTimeoutSeconds: service.shutdownTimeoutSeconds,
		docs: { basicAuth },
	};
	return { app: raw.app, [runtime]: selected };
}
export function loadRuntimeConfig<R extends RuntimeName>(runtime: R, options?: ConfigOptions): RuntimeConfig<R>;
export function loadRuntimeConfig(runtime: RuntimeName, options: ConfigOptions = {}): RuntimeConfig<RuntimeName> {
	const input = snapshot(options, runtime);
	return parse(runtimeSchemas[runtime], selectRuntime(runtime, input), input, runtime);
}
let processSnapshot: Snapshot | undefined;
function cached<T>(runtime: string, load: (input: Snapshot) => T): () => T {
	let value: { result: T } | undefined;
	return () => (value ??= { result: load((processSnapshot ??= snapshot({}, runtime))) }).result;
}
export const getWebConfig = cached("web", (input) => parse(runtimeSchemas.web, selectRuntime("web", input), input, "web"));
export const getApiConfig = cached("api", (input) => parse(runtimeSchemas.api, selectRuntime("api", input), input, "api"));
function webToolConfig(input: Snapshot) {
	return parse(webToolConfigSchema, { app: input.raw.app, web: { host: input.raw.web.host, port: input.raw.web.port } }, input, "web-tool");
}
export const getWebToolConfig = cached("web-tool", webToolConfig);
export const loadWebToolConfig = (options: ConfigOptions = {}) => webToolConfig(snapshot(options, "web-tool"));
export const getBOConfig = cached("bo", (input) => parse(runtimeSchemas.bo, selectRuntime("bo", input), input, "bo"));
function boToolConfig(input: Snapshot) {
	return parse(boToolConfigSchema, { app: input.raw.app, bo: { host: input.raw.bo.host, port: input.raw.bo.port } }, input, "bo-tool");
}
export const getBOToolConfig = cached("bo-tool", boToolConfig);
export const loadBOToolConfig = (options: ConfigOptions = {}) => boToolConfig(snapshot(options, "bo-tool"));
function publicBOConfig(input: Snapshot) {
	return parse(publicBOConfigSchema, { boApiBaseUrl: input.raw.bo.boApiBaseUrl }, input, "bo-public");
}
export const getPublicBOConfig = cached("bo-public", publicBOConfig);
export const loadPublicBOConfig = (options: ConfigOptions = {}) => publicBOConfig(snapshot(options, "bo-public"));
function publicWebConfig(input: Snapshot) {
	return parse(publicWebConfigSchema, { apiBaseUrl: input.raw.web.apiBaseUrl }, input, "web-public");
}
export const getPublicWebConfig = cached("web-public", publicWebConfig);
export const loadPublicWebConfig = (options: ConfigOptions = {}) => publicWebConfig(snapshot(options, "web-public"));

