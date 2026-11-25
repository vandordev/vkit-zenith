//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOCustomerActivityExactSafeScope(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		org, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Workspace", Slug: "workspace", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		other, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Other", Slug: "other", OwnerUserID: org.User.ID})
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "before@example.test", DisplayName: "Before"})
		if err != nil {
			return err
		}
		u, err = c.UpdateUser(ctx, a, u.ID, customers.UserProfile{DisplayName: "Current label", Revision: customers.UserRevision(u)})
		if err != nil {
			return err
		}
		marker := "fixture-secret-marker-do-not-project"
		confirmed := func(rev string) customers.Sensitive {
			return customers.Sensitive{Revision: rev, Reason: marker, CurrentPassword: fixturePassword}
		}
		u, err = c.VerifyUserEmail(ctx, a, u.ID, confirmed(customers.UserRevision(u)))
		if err != nil {
			return err
		}
		u, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: confirmed(customers.UserRevision(u)), Email: "after@example.test"})
		if err != nil {
			return err
		}
		u, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: confirmed(customers.UserRevision(u)), Status: "disabled"})
		if err != nil {
			return err
		}
		u, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: confirmed(customers.UserRevision(u)), Status: "active"})
		if err != nil {
			return err
		}
		_, err = c.UpdateOrganization(ctx, a, org.Organization.ID, customers.OrganizationProfile{Name: "Renamed workspace", Slug: "workspace", Revision: customers.OrganizationRevision(org.Organization)})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, a, org.Organization.ID, u.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		m, err = c.SetRole(ctx, a, org.Organization.ID, m.ID, customers.RoleChange{Sensitive: confirmed(customers.MembershipRevision(m)), Role: customers.RoleOwner})
		if err != nil {
			return err
		}
		m, err = c.SetRole(ctx, a, org.Organization.ID, m.ID, customers.RoleChange{Sensitive: confirmed(customers.MembershipRevision(m)), Role: customers.RoleMember})
		if err != nil {
			return err
		}
		m, err = c.RemoveMember(ctx, a, org.Organization.ID, m.ID, customers.MembershipRevision(m))
		if err != nil {
			return err
		}
		_, err = c.RestoreMember(ctx, a, org.Organization.ID, m.ID, customers.MembershipRevision(m))
		if err != nil {
			return err
		}
		_, err = c.SetOrganizationStatus(ctx, a, other.Organization.ID, customers.OrganizationStatus{Sensitive: confirmed(customers.OrganizationRevision(other.Organization)), Status: "suspended"})
		if err != nil {
			return err
		}
		adminIn := bootstrapInput("activity-admin@example.test")
		adminIn.Role = "admin"
		adminIn.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, adminIn)
		if err != nil {
			return err
		}
		// Valid append-only historical storage with malformed/unknown IDs and wrong
		// org attribution must not become trusted navigation merely by UUID parsing.
		malformedID := "00000000-0000-4000-8000-000000000081"
		wrongScopeID := "00000000-0000-4000-8000-000000000082"
		unknownTargetID := "00000000-0000-4000-8000-000000000083"
		if _, err = p.Exec(ctx, `INSERT INTO audit_events(id,occurred_at,actor_kind,actor_staff_id,organization_id,action,outcome,target_type,target_id,detail) VALUES
   ($6,$1,'staff',$2,NULL,'customer.user.update','success','user','not-a-uuid','{"changed_fields":["display_name"]}'),
   (gen_random_uuid(),$1,'staff',$2,NULL,'customer.unknown','success','user',$3,'{}'),
   ($7,$1,'staff',$2,$4,'customer.membership.role','success','organization_membership',$5,'{"changed_fields":["role"]}'),
   ($8,$1,'staff',$2,NULL,'customer.user.update','success','unknown_historical_target',$3,'{"changed_fields":["display_name"]}')`, fixtureNow, a.StaffID, u.ID.String(), other.Organization.ID, m.ID.String(), malformedID, wrongScopeID, unknownTargetID); err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, _, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		adminCookies, _, err := loginHTTP(e, admin.Email)
		if err != nil {
			return err
		}
		get := func(path string, cookies []*http.Cookie) ([]byte, error) {
			w := perform(e, "GET", path, "", boOrigin, "", cookies)
			if w.Code != 200 {
				return nil, fmt.Errorf("%s: %d %s", path, w.Code, w.Body.String())
			}
			for _, bad := range []string{marker, fixturePassword, "before@example.test", "after@example.test", "\"reason\"", "\"detail\"", "tokenHash"} {
				if strings.Contains(w.Body.String(), bad) {
					return nil, fmt.Errorf("unsafe activity projection %s at %s", bad, path)
				}
			}
			return w.Body.Bytes(), nil
		}
		actions := []string{"customer.user.create", "customer.user.update", "customer.user.status", "customer.user.email_change", "customer.user.email_verify", "customer.organization.create", "customer.organization.update", "customer.organization.status", "customer.membership.add", "customer.membership.remove", "customer.membership.restore", "customer.membership.role"}
		for _, action := range actions {
			raw, err := get("/v1/activity?action="+action+"&pageSize=100", cookies)
			if err != nil {
				return err
			}
			var list v1.V1ActivityListBody
			if err = json.Unmarshal(raw, &list); err != nil {
				return err
			}
			if len(list.Items) == 0 || int64(len(list.Items)) != list.Pagination.Total {
				return fmt.Errorf("missing/miscounted activity %s", action)
			}
			resolved := false
			for _, row := range list.Items {
				raw, err = get("/v1/activity/"+row.ID, cookies)
				if err != nil {
					return err
				}
				var detail v1.V1ActivityDetail
				if err = json.Unmarshal(raw, &detail); err != nil {
					return err
				}
				if len(detail.ChangedFields) == 0 || detail.Previous != nil || detail.Current != nil {
					return fmt.Errorf("customer changes misprojected %s", action)
				}
				if row.ID == malformedID || row.ID == wrongScopeID || row.ID == unknownTargetID {
					if row.Target.Available || detail.Target.Available || row.Target.OrganizationID != nil || detail.Target.OrganizationID != nil {
						return fmt.Errorf("historical invalid target acquired navigation: %s", raw)
					}
				}
				if row.Target.Available {
					resolved = true
					if row.Target.ID == nil || row.Target.Name == nil {
						return fmt.Errorf("missing resolved target")
					}
					if row.Target.Type != nil && *row.Target.Type == "user" && *row.Target.ID == u.ID.String() && *row.Target.Name != "Current label" {
						return fmt.Errorf("user target is not the current display label")
					}
					if row.Target.Type != nil && *row.Target.Type == "organization" && *row.Target.ID == org.Organization.ID.String() && *row.Target.Name != "Renamed workspace" {
						return fmt.Errorf("organization target is not the current display label")
					}
					if strings.HasPrefix(action, "customer.membership.") {
						expected := map[string]string{org.Owner.ID.String(): org.Organization.ID.String(), other.Owner.ID.String(): other.Organization.ID.String(), m.ID.String(): org.Organization.ID.String()}
						if detail.Target.OrganizationID == nil || *detail.Target.OrganizationID != expected[*row.Target.ID] || row.Target.OrganizationID == nil || *row.Target.OrganizationID != *detail.Target.OrganizationID {
							return fmt.Errorf("membership lacks authoritative organization: %s", raw)
						}
					}
				} else {
					if detail.Target.OrganizationID != nil || row.Target.OrganizationID != nil {
						return fmt.Errorf("unavailable target acquired organization")
					}
				}
			}
			if !resolved {
				return fmt.Errorf("no navigable current target for %s", action)
			}
		}
		for _, path := range []string{"/v1/activity/actions", "/v1/activity/recent", "/v1/activity/staff-choices"} {
			raw, err := get(path, cookies)
			if err != nil {
				return err
			}
			if path == "/v1/activity/actions" {
				for _, action := range actions {
					if !strings.Contains(string(raw), action) {
						return fmt.Errorf("missing action choice %s", action)
					}
				}
			}
			if path == "/v1/activity/recent" {
				var recent v1.V1ActivityRecentBody
				if err = json.Unmarshal(raw, &recent); err != nil {
					return err
				}
				if len(recent.Items) != 5 {
					return fmt.Errorf("recent bound")
				}
			}
			if w := perform(e, "GET", path, "", boOrigin, "", adminCookies); w.Code != 403 {
				return fmt.Errorf("Admin activity allowed %s: %d", path, w.Code)
			}
		}
		raw, err := get("/v1/activity?pageSize=100", cookies)
		if err != nil {
			return err
		}
		if strings.Contains(string(raw), "customer.unknown") {
			return fmt.Errorf("unknown action admitted")
		}
		if !strings.Contains(string(raw), "staff.bootstrap") {
			return fmt.Errorf("staff compatibility lost")
		}
		var all v1.V1ActivityListBody
		if err = json.Unmarshal(raw, &all); err != nil {
			return err
		}
		if len(all.Items) == 0 {
			return fmt.Errorf("empty activity")
		}
		for _, path := range []string{"/v1/activity", "/v1/activity/" + all.Items[0].ID} {
			if w := perform(e, "GET", path, "", boOrigin, "", adminCookies); w.Code != 403 {
				return fmt.Errorf("Admin activity leaked %s: %d", path, w.Code)
			}
		}
		return nil
	})
}
