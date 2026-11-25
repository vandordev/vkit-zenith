export type FormStatus =
	| { kind: "idle" }
	| { kind: "pending" }
	| { kind: "error"; message: string; fields?: Partial<Record<string, string>> }
	| { kind: "rate-limited"; retryAfterSeconds?: number }
	| { kind: "unavailable" };
