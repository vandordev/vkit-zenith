import type { BoV1GetMeResponse } from "../lib/api/generated/types.gen";

// Deterministic presentation fixtures only. Never import into runtime adapters.
export const id = "11111111-1111-4111-8111-111111111111";
export const other = "22222222-2222-4222-8222-222222222222";
export const revision = "a".repeat(64);
export const time = "2026-10-25T08:00:00Z";
export const noop = () => {};
export const staff: BoV1GetMeResponse = {
	id,
	email: "staf@example.invalid",
	displayName: "Nadia Prameswari",
	role: "superadmin",
	status: "active",
	updatedAt: time,
	revision,
};
