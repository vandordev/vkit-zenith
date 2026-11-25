package fixtures

import "github.com/google/uuid"

const fixtureCreated = "2026-11-01T00:00:00Z"

func UserSessions(userID uuid.UUID) []UserSession {
	revoked := "2026-11-03T00:00:00Z"
	return []UserSession{
		{ID: FixtureID(userID, "session-active").String(), AuthenticatedAt: fixtureCreated, LastSeenAt: "2026-11-02T00:00:00Z", ExpiresAt: "2026-12-20T00:00:00Z", CreatedAt: fixtureCreated},
		{ID: FixtureID(userID, "session-revoked").String(), AuthenticatedAt: fixtureCreated, LastSeenAt: "2026-11-02T00:00:00Z", ExpiresAt: "2026-12-20T00:00:00Z", RevokedAt: &revoked, CreatedAt: fixtureCreated},
	}
}

func UserEmailActions(userID uuid.UUID) []EmailAction {
	consumed, revoked := "2026-11-01T01:00:00Z", "2026-11-01T02:00:00Z"
	return []EmailAction{
		{ID: FixtureID(userID, "email-action-verify").String(), Purpose: "verify_email", CreatedAt: fixtureCreated, ExpiresAt: "2026-11-02T00:00:00Z", ConsumedAt: &consumed},
		{ID: FixtureID(userID, "email-action-reset").String(), Purpose: "reset_password", CreatedAt: fixtureCreated, ExpiresAt: "2026-11-02T00:00:00Z", RevokedAt: &revoked},
		{ID: FixtureID(userID, "email-action-setup").String(), Purpose: "setup_password", CreatedAt: "2026-11-03T00:00:00Z", ExpiresAt: "2026-11-04T00:00:00Z"},
		{ID: FixtureID(userID, "email-action-verify-pending").String(), Purpose: "verify_email", CreatedAt: "2026-11-03T01:00:00Z", ExpiresAt: "2026-11-04T01:00:00Z"},
	}
}

func StaffSessions(staffID uuid.UUID) []StaffSession {
	revoked := "2026-11-03T00:00:00Z"
	return []StaffSession{
		{ID: FixtureID(staffID, "staff-session-active").String(), CreatedAt: fixtureCreated, ExpiresAt: "2026-12-01T00:00:00Z"},
		{ID: FixtureID(staffID, "staff-session-revoked").String(), CreatedAt: fixtureCreated, ExpiresAt: "2026-12-01T00:00:00Z", RevokedAt: &revoked},
	}
}

func UserHistory(userID uuid.UUID) []History {
	targetType, targetID := "user", userID.String()
	staffID := FixtureID(userID, "history-staff").String()
	system := "identity.email_verification"
	return []History{
		{ID: FixtureID(userID, "user-history-created").String(), OccurredAt: fixtureCreated, ActorKind: "system", ActorName: "Sistem", System: &system, Action: "user.created", Outcome: "success", TargetType: &targetType, TargetID: &targetID},
		{ID: FixtureID(userID, "user-history-login").String(), OccurredAt: "2026-11-02T00:00:00Z", ActorKind: "staff", ActorID: &staffID, ActorName: "Staf Fixture", Role: stringPtr("admin"), Action: "user.session_revoked", Outcome: "success", TargetType: &targetType, TargetID: &targetID, Reason: stringPtr("Menutup sesi yang tidak dikenal")},
	}
}

func StaffHistory(staffID uuid.UUID) []History {
	targetType, targetID := "platform_staff", staffID.String()
	actorID := FixtureID(staffID, "history-actor").String()
	system := "staff.bootstrap"
	return []History{
		{ID: FixtureID(staffID, "staff-history-created").String(), OccurredAt: fixtureCreated, ActorKind: "system", ActorName: "Sistem", System: &system, Action: "staff.created", Outcome: "success", TargetType: &targetType, TargetID: &targetID},
		{ID: FixtureID(staffID, "staff-history-updated").String(), OccurredAt: "2026-11-02T00:00:00Z", ActorKind: "staff", ActorID: &actorID, ActorName: "Superadmin Fixture", Role: stringPtr("superadmin"), Action: "staff.updated", Outcome: "success", TargetType: &targetType, TargetID: &targetID, Reason: stringPtr("Memperbarui profil staf")},
	}
}

func stringPtr(value string) *string { return &value }
