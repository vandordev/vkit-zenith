package v1

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type activityPolicy struct {
	action, label, target string
	fields                []string
}

// Exact installed emitters, not a wildcard grant to future audit actions.
var activityPolicies = []activityPolicy{
	{"staff.bootstrap", "Inisialisasi staf", "staff", []string{"email_normalized", "display_name", "role", "status", "credentials"}},
	{"staff.create", "Staf dibuat", "staff", []string{"email_normalized", "display_name", "role", "status", "credentials"}},
	{"staff.update", "Staf diperbarui", "staff", []string{"email_normalized", "display_name", "role", "status"}},
	{"staff.password_replace", "Kata sandi staf diganti", "staff", []string{"credentials"}},
	{"staff.password_change", "Kata sandi sendiri diganti", "staff", []string{"credentials"}},
	{"staff.login", "Masuk staf", "staff", nil},
	{"staff.logout", "Keluar staf", "staff", nil},
	{"staff.confirmation", "Konfirmasi staf", "auth_bucket", nil},
	{"customer.user.create", "Pengguna pelanggan dibuat", "user", []string{"email", "display_name", "status"}},
	{"customer.user.update", "Profil pelanggan diperbarui", "user", []string{"display_name"}},
	{"customer.user.status", "Status pelanggan diubah", "user", []string{"status"}},
	{"customer.user.email_change", "Email pelanggan diganti", "user", []string{"email", "email_verified_at"}},
	{"customer.user.email_verify", "Email pelanggan diverifikasi manual", "user", []string{"email_verified_at"}},
	{"customer.organization.create", "Organisasi dibuat", "organization", []string{"name", "slug", "status"}},
	{"customer.organization.update", "Profil organisasi diperbarui", "organization", []string{"name", "slug"}},
	{"customer.organization.status", "Status organisasi diubah", "organization", []string{"status"}},
	{"customer.membership.add", "Anggota organisasi ditambahkan", "organization_membership", []string{"status"}},
	{"customer.membership.remove", "Keanggotaan dihapus", "organization_membership", []string{"status"}},
	{"customer.membership.restore", "Keanggotaan dipulihkan", "organization_membership", []string{"status"}},
	{"customer.membership.role", "Kepemilikan organisasi diubah", "organization_membership", []string{"role"}},
}

func activityPolicyFor(action string) (activityPolicy, bool) {
	for _, p := range activityPolicies {
		if p.action == action {
			return p, true
		}
	}
	return activityPolicy{}, false
}
func activityActions() []string {
	out := make([]string, 0, len(activityPolicies))
	for _, p := range activityPolicies {
		out = append(out, p.action)
	}
	return out
}

// Internal input only: raw detail and historical target IDs never serialize.
type activityRecord struct {
	id                               pgtype.UUID
	occurredAt                       time.Time
	action, outcome, actorKind       string
	actorID                          pgtype.UUID
	actorName                        pgtype.Text
	targetType, targetID, targetName pgtype.Text
	targetOrganizationID             pgtype.UUID
	targetAvailable                  bool
	detail                           []byte
}

func activityText(v pgtype.Text) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}
func activityUUID(v pgtype.UUID) *string {
	if v.Valid {
		s := v.String()
		return &s
	}
	return nil
}
func projectActivity(r activityRecord) V1ActivityRow {
	p, _ := activityPolicyFor(r.action)
	out := V1ActivityRow{ID: r.id.String(), OccurredAt: r.occurredAt.UTC(), Action: r.action, Outcome: r.outcome,
		Actor: V1ActivityActor{Kind: r.actorKind}}
	if r.actorKind == "staff" {
		out.Actor.ID = activityUUID(r.actorID)
		out.Actor.Name = activityText(r.actorName)
	}
	// Login failures and confirmations are anonymous summaries. Never surface
	// the auth_bucket target's source/account hashes, even as an unavailable ID.
	if r.targetType.String == "auth_bucket" {
		typeName := "auth_summary"
		out.Target.Type = &typeName
	} else if r.targetType.Valid && r.targetType.String == p.target && p.target != "auth_bucket" {
		out.Target.Type = &p.target
		var id pgtype.UUID
		if id.Scan(r.targetID.String) == nil && id.Valid {
			out.Target.ID = activityUUID(id)
			out.Target.Name = activityText(r.targetName)
			out.Target.Available = r.targetAvailable
			if p.target == "organization_membership" {
				out.Target.Available = out.Target.Available && r.targetOrganizationID.Valid
				if out.Target.Available {
					out.Target.OrganizationID = activityUUID(r.targetOrganizationID)
				}
			}
		}
	}
	return out
}

func projectActivityDetail(r activityRecord) V1ActivityDetail {
	out := V1ActivityDetail{V1ActivityRow: projectActivity(r), ChangedFields: []string{}}
	p, ok := activityPolicyFor(r.action)
	if !ok || len(r.detail) > 65536 || r.targetType.String == "auth_bucket" {
		return out
	}
	// Decode only known structural slots; free-text values and unknown keys have
	// no output destination. Invalid structure means unavailable, not invented.
	var d struct {
		ChangedFields []string `json:"changed_fields"`
		Previous      *struct {
			Role   string `json:"role"`
			Status string `json:"status"`
		} `json:"previous"`
		Current *struct {
			Role   string `json:"role"`
			Status string `json:"status"`
		} `json:"current"`
	}
	if json.Unmarshal(r.detail, &d) != nil || len(d.ChangedFields) > 100 {
		return out
	}
	for _, key := range d.ChangedFields {
		if slices.Contains(p.fields, key) && !slices.Contains(out.ChangedFields, key) {
			out.ChangedFields = append(out.ChangedFields, key)
		}
	}
	if r.action == "staff.update" {
		state := func(role, status string) *V1ActivityStaffState {
			s := &V1ActivityStaffState{}
			if role == "superadmin" || role == "admin" {
				s.Role = &role
			}
			if status == "active" || status == "disabled" {
				s.Status = &status
			}
			if s.Role == nil && s.Status == nil {
				return nil
			}
			return s
		}
		if d.Previous != nil {
			out.Previous = state(d.Previous.Role, d.Previous.Status)
		}
		if d.Current != nil {
			out.Current = state(d.Current.Role, d.Current.Status)
		}
	}
	out.DetailAvailable = len(out.ChangedFields) > 0 || out.Previous != nil || out.Current != nil
	return out
}
