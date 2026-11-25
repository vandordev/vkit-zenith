//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

// BOCO-CUST-002: read availability must permit the same explicit disabled-owner
// demotion as the protected command, without granting disabled-user promotion.
func TestBOCustomerReadDisabledOwnerDemotionThenHTTPRemoval(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		org, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Owner maintenance", Slug: "owner-maintenance", NewOwner: &customers.UserCreate{Email: "first@example.test", DisplayName: "First owner"}})
		if err != nil {
			return err
		}
		confirmed := func(rev string) customers.Sensitive {
			return customers.Sensitive{Revision: rev, Reason: "Isolated owner maintenance", CurrentPassword: fixturePassword}
		}
		replacement, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "replacement@example.test", DisplayName: "Replacement"})
		if err != nil {
			return err
		}
		replacementMember, err := c.AddMember(ctx, a, org.Organization.ID, replacement.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		if _, err = c.SetRole(ctx, a, org.Organization.ID, replacementMember.ID, customers.RoleChange{Sensitive: confirmed(customers.MembershipRevision(replacementMember)), Role: "owner"}); err != nil {
			return err
		}
		if _, err = c.SetUserStatus(ctx, a, org.User.ID, customers.UserStatus{Sensitive: confirmed(customers.UserRevision(org.User)), Status: "disabled"}); err != nil {
			return err
		}
		nonowner, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "nonowner@example.test", DisplayName: "Nonowner"})
		if err != nil {
			return err
		}
		nonownerMember, err := c.AddMember(ctx, a, org.Organization.ID, nonowner.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		if _, err = c.SetUserStatus(ctx, a, nonowner.ID, customers.UserStatus{Sensitive: confirmed(customers.UserRevision(nonowner)), Status: "disabled"}); err != nil {
			return err
		}
		removedUser, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "removed@example.test", DisplayName: "Removed"})
		if err != nil {
			return err
		}
		removed, err := c.AddMember(ctx, a, org.Organization.ID, removedUser.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		if _, err = c.RemoveMember(ctx, a, org.Organization.ID, removed.ID, customers.MembershipRevision(removed)); err != nil {
			return err
		}
		in := bootstrapInput("owner-admin@example.test")
		in.Role, in.CurrentPassword = "admin", fixturePassword
		admin, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		rootCookies, rootCSRF, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		adminCookies, adminCSRF, err := loginHTTP(e, admin.Email)
		if err != nil {
			return err
		}
		base := "/v1/organizations/" + org.Organization.ID.String() + "/memberships"
		get := func(id string, cookies []*http.Cookie) (v1.V1OrganizationMembershipGetBody, error) {
			var out v1.V1OrganizationMembershipGetBody
			w := perform(e, "GET", base+"/"+id, "", boOrigin, "", cookies)
			if w.Code != 200 {
				return out, fmt.Errorf("member read: %d %s", w.Code, w.Body.String())
			}
			return out, json.Unmarshal(w.Body.Bytes(), &out)
		}
		owner, err := get(org.Owner.ID.String(), rootCookies)
		if err != nil {
			return err
		}
		if !owner.CanChangeRole || owner.RoleBlockedReason != "" || owner.CanRemove || owner.UserStatus != "disabled" || owner.Role != "owner" {
			return fmt.Errorf("disabled owner with replacement cannot explicitly demote: %+v", owner)
		}
		w := perform(e, "GET", base, "", boOrigin, "", rootCookies)
		var list v1.V1OrganizationMembershipListBody
		if w.Code != 200 {
			return fmt.Errorf("member list: %d", w.Code)
		}
		if err = json.Unmarshal(w.Body.Bytes(), &list); err != nil {
			return err
		}
		found := false
		for _, row := range list.Items {
			if row.ID == owner.ID {
				found = true
				if !row.CanChangeRole || row.RoleBlockedReason != "" {
					return fmt.Errorf("list disagrees with detail availability")
				}
			}
		}
		if !found {
			return fmt.Errorf("owner absent from list")
		}
		for _, check := range []struct {
			id      string
			cookies []*http.Cookie
		}{
			{nonownerMember.ID.String(), rootCookies}, {removed.ID.String(), rootCookies},
			{replacementMember.ID.String(), rootCookies}, // Sole eligible owner cannot be revoked.
			{org.Owner.ID.String(), adminCookies},
		} {
			detail, err := get(check.id, check.cookies)
			if err != nil {
				return err
			}
			if detail.CanChangeRole || detail.RoleBlockedReason == "" {
				return fmt.Errorf("restricted owner action became available: %+v", detail)
			}
		}
		body := map[string]any{"revision": owner.Revision, "reason": "Explicit demotion", "currentPassword": fixturePassword, "role": "member"}
		w = perform(e, "POST", base+"/"+owner.ID+"/role", customerJSON(body), boOrigin, rootCSRF, rootCookies)
		if w.Code != 200 {
			return fmt.Errorf("protected demotion: %d %s", w.Code, w.Body.String())
		}
		demoted, err := get(owner.ID, adminCookies)
		if err != nil {
			return err
		}
		if demoted.Role == "owner" || !demoted.CanRemove || demoted.RemoveBlockedReason != "" {
			return fmt.Errorf("explicitly demoted membership cannot remove")
		}
		w = perform(e, "POST", base+"/"+owner.ID+"/remove", customerJSON(map[string]string{"revision": demoted.Revision}), boOrigin, adminCSRF, adminCookies)
		if w.Code != 200 {
			return fmt.Errorf("ordinary removal: %d %s", w.Code, w.Body.String())
		}
		final, err := get(owner.ID, rootCookies)
		if err != nil {
			return err
		}
		if final.ID != owner.ID || final.Status != "removed" || final.Role == "owner" || final.UserStatus != "disabled" || final.CanChangeRole {
			return fmt.Errorf("incorrect final membership: %+v", final)
		}
		return nil
	})
}
