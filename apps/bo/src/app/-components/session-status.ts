/** Derived presentation status of login sessions and email-action tokens; the API stores only timestamps. */
type Timestamps = { expiresAt: string; revokedAt: string | null };

export const sessionStatus = ({ expiresAt, revokedAt }: Timestamps, now: number): "revoked" | "expired" | "active" =>
	revokedAt ? "revoked" : Date.parse(expiresAt) <= now ? "expired" : "active";

export const emailActionStatus = (action: Timestamps & { consumedAt: string | null }, now: number): "consumed" | "revoked" | "expired" | "pending" =>
	action.consumedAt ? "consumed" : action.revokedAt ? "revoked" : Date.parse(action.expiresAt) <= now ? "expired" : "pending";
