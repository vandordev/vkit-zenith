import type { AuthTransport } from "./client";
import { AuthError } from "./errors";
import {
	webV1CreateApiKey,
	webV1GetOrganizationProfile,
	webV1ListApiKeys,
	webV1ListMyOrganizations,
	webV1ListOrganizationMembers,
	webV1RevokeApiKey,
	webV1UploadLogo,
} from "./generated/sdk.gen";
import type { ApiKeyCreateInputBody, LogoUploadInputBody, WebV1ListApiKeysData } from "./generated/types.gen";

const read = (t: AuthTransport) => ({ client: t.client, throwOnError: true as const });
const sessionChanged = () => new AuthError(401, "session_changed", "Sesi berubah.");

/** Rejects when the identity changed while the CSRF token was awaited or while the call ran, so a stale token or result never escapes. */
export async function orgWrite<T>(t: AuthTransport, call: (headers: { "X-CSRF-Token": string }) => Promise<T>): Promise<T> {
	const accepted = t.epoch();
	const token = await t.csrf();
	if (t.epoch() !== accepted) throw sessionChanged();
	const result = await call({ "X-CSRF-Token": token });
	if (t.epoch() !== accepted) throw sessionChanged();
	return result;
}

export const listMyOrganizations = async (t: AuthTransport) => (await webV1ListMyOrganizations(read(t))).data.items;
export const getProfile = async (t: AuthTransport, organizationId: string) =>
	(await webV1GetOrganizationProfile({ ...read(t), path: { organizationId } })).data;
export const listMembers = async (t: AuthTransport, organizationId: string) =>
	(await webV1ListOrganizationMembers({ ...read(t), path: { organizationId } })).data.items;

// Settings lists send only the filters that are set; dummy writes return the `result`, persisted nowhere.
export type SettingsQuery<S extends string> = { q?: string; status?: S; page?: number; size?: number };
const settingsListQuery = <S extends string>({ q, status, page, size }: SettingsQuery<S> = {}) => ({
	...(q && { q }),
	...(status && { status }),
	...(page && page > 1 && { page }),
	...(size && { size }),
});
export const listApiKeys = async (
	t: AuthTransport,
	organizationId: string,
	query?: SettingsQuery<NonNullable<NonNullable<WebV1ListApiKeysData["query"]>["status"]>>,
) => (await webV1ListApiKeys({ ...read(t), path: { organizationId }, query: settingsListQuery(query) })).data;
/** The result carries the one-time `secret`. */
export const createApiKey = (t: AuthTransport, organizationId: string, body: ApiKeyCreateInputBody) =>
	orgWrite(t, async (headers) => (await webV1CreateApiKey({ ...read(t), path: { organizationId }, headers, body })).data.result);
export const revokeApiKey = (t: AuthTransport, organizationId: string, keyId: string) =>
	orgWrite(t, async (headers) => (await webV1RevokeApiKey({ ...read(t), path: { organizationId, keyId }, headers })).data.result);
/** Accepts a content type and size only, no file. */
export const uploadLogo = (t: AuthTransport, organizationId: string, body: LogoUploadInputBody) =>
	orgWrite(t, async (headers) => (await webV1UploadLogo({ ...read(t), path: { organizationId }, headers, body })).data.result);

export * from "../organization-pure";
