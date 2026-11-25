import { describe, expect, test } from "bun:test";
import { emailActionStatus, sessionStatus } from "./session-status";

const now = Date.parse("2026-11-01T00:00:00Z");
const live = "2026-12-01T00:00:00Z";
const past = "2026-10-01T00:00:00Z";
describe("sessionStatus", () => {
	test("revoked wins over expiry, expiry over active", () => {
		expect(sessionStatus({ expiresAt: live, revokedAt: past }, now)).toBe("revoked");
		expect(sessionStatus({ expiresAt: past, revokedAt: null }, now)).toBe("expired");
		expect(sessionStatus({ expiresAt: live, revokedAt: null }, now)).toBe("active");
	});
});
describe("emailActionStatus", () => {
	test("consumed, revoked, expired and pending are distinguished from the timestamps alone", () => {
		expect(emailActionStatus({ expiresAt: live, revokedAt: null, consumedAt: past }, now)).toBe("consumed");
		expect(emailActionStatus({ expiresAt: live, revokedAt: past, consumedAt: null }, now)).toBe("revoked");
		expect(emailActionStatus({ expiresAt: past, revokedAt: null, consumedAt: null }, now)).toBe("expired");
		expect(emailActionStatus({ expiresAt: live, revokedAt: null, consumedAt: null }, now)).toBe("pending");
	});
});
