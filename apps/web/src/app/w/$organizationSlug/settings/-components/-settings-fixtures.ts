import type { ApiKey, V1CreateApiKeyResult } from "../../../../../lib/api/generated/types.gen";

// Deterministic, fake, non-sensitive settings fixtures shaped like the generated Web API DTOs.
/** The API's unusable dummy constant; a real key is never in a fixture. */
export const dummySecret = "zk_DUMMYNOTAKEY";
const id = (kind: string, n: number) => `0f3a6c1e-${kind}-4000-8000-${String(n).padStart(12, "0")}`;

export const apiKeys: ApiKey[] = [
	{
		id: id("0d01", 1),
		name: "Kunci integrasi toko",
		prefix: "zk_abcd",
		status: "active",
		lastUsedAt: "2026-11-18T08:30:00Z",
		expiresAt: "2026-12-31T16:59:59Z",
		createdAt: "2026-10-01T00:00:00Z",
	},
	{
		id: id("0d01", 2),
		name: "Kunci baru belum dipakai",
		prefix: "zk_new0",
		status: "active",
		createdAt: "2026-11-17T00:00:00Z",
	},
	{
		id: id("0d01", 3),
		name: "Kunci kedaluwarsa",
		prefix: "zk_old1",
		status: "expired",
		expiresAt: "2026-11-01T00:00:00Z",
		lastUsedAt: "2026-10-30T00:00:00Z",
		createdAt: "2026-09-01T00:00:00Z",
	},
	{
		id: id("0d01", 4),
		name: "Kunci lama",
		prefix: "zk_old0",
		status: "revoked",
		revokedAt: "2026-11-02T00:00:00Z",
		createdAt: "2026-09-01T00:00:00Z",
	},
];
export const longApiKey: ApiKey = {
	...apiKeys[0]!,
	id: id("0d01", 9),
	name: "Kunci integrasi untuk sinkronisasi stok gudang pusat Perkumpulan Penggiat Kopi Nusantara Tenggara Timur",
};
export const createdApiKey = (): V1CreateApiKeyResult => ({
	key: { id: "simulated-api-key", name: "Kunci integrasi baru", prefix: "zk_dummy", status: "active", createdAt: "2026-11-19T00:00:00Z" },
	secret: dummySecret,
});
