//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

// Catches filtered totals replacing structural summaries, fanout counts,
// wildcard search, incomplete selected context and leaked customer secrets.
func TestBOCustomerReadJoinedScopeAndSelectors(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		org, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Org_100%", Slug: "org-one", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		other, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Other", Slug: "org-two", OwnerUserID: org.User.ID})
		if err != nil {
			return err
		}
		member, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "member@example.test", DisplayName: "Member"})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, a, org.Organization.ID, member.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		if _, err = c.RemoveMember(ctx, a, org.Organization.ID, m.ID, customers.MembershipRevision(m)); err != nil {
			return err
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='disabled',email_verified_at=$2 WHERE id=$1`, member.ID, fixtureNow); err != nil {
			return err
		}
		if _, err = p.Exec(ctx, `UPDATE organizations SET status='suspended' WHERE id=$1`, other.Organization.ID); err != nil {
			return err
		}
		// More than one default page, without creating credentials or memberships.
		if _, err = p.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) SELECT md5('customer-read-fixture-'||n)::uuid,'extra'||n||'@example.test','extra'||n||'@example.test','Extra '||n,'active',$1,$1 FROM generate_series(1,26) n`, fixtureNow); err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, _, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		checks := []struct {
			path      string
			status    int
			fragments []string
		}{
			{"/v1/users", 200, []string{`"total":"28"`, `"active":"27"`, `"disabled":"1"`, `"verified":"1"`, `"unverified":"27"`}},
			{"/v1/users?status=disabled", 200, []string{`"total":"1"`, `"active":"27"`, `"emailVerifiedAt":"`}},
			{"/v1/users?page=99", 200, []string{`"items":[]`, `"total":"28"`}},
			{"/v1/users/" + org.User.ID.String(), 200, []string{`"membershipCount":"2"`, `"activeMembershipCount":"2"`, `"ownerMembershipCount":"2"`, `"revision":"`}},
			{"/v1/users/" + member.ID.String() + "/memberships", 200, []string{`"status":"removed"`, `"organizationName":"Org_100%"`}},
			{"/v1/organizations?q=%25", 200, []string{`"name":"Org_100%"`, `"total":"1"`, `"suspended":"1"`}},
			{"/v1/organizations?q=_", 200, []string{`"name":"Org_100%"`, `"total":"1"`}},
			{"/v1/organizations/" + org.Organization.ID.String(), 200, []string{`"memberCount":"2"`, `"activeMemberCount":"1"`, `"activeOwnerCount":"1"`}},
			{"/v1/organizations/" + org.Organization.ID.String() + "/memberships?status=removed", 200, []string{`"total":"1"`, `"active":"1"`, `"removed":"1"`, `"owners":"1"`, `"userStatus":"disabled"`}},
			{"/v1/organizations/" + other.Organization.ID.String() + "/memberships/" + m.ID.String(), 404, nil},
			{"/v1/users/choices?q=absent&selectedId=" + member.ID.String(), 200, []string{`"items":[]`, `"selected":{`, `"eligible":false`, `"id":"` + member.ID.String() + `"`}},
			{"/v1/users/choices?organizationId=" + org.Organization.ID.String() + "&selectedId=" + org.User.ID.String(), 200, []string{`"eligible":false`, `"disabledReason":"`}},
			{"/v1/users/choices?organizationId=" + org.Organization.ID.String() + "&selectedId=" + member.ID.String(), 200, []string{`"eligible":false`}},
			{"/v1/users?verification=wrong", 422, nil}, {"/v1/organizations?status=wrong", 422, nil}, {"/v1/users/not-a-uuid", 422, nil},
			{"/v1/users/choices?selectedId=wrong", 422, nil}, {"/v1/organizations/" + org.Organization.ID.String() + "/memberships?ownership=wrong", 422, nil},
		}
		for _, check := range checks {
			w := perform(e, "GET", check.path, "", boOrigin, "", cookies)
			if w.Code != check.status {
				return fmt.Errorf("%s: status %d: %s", check.path, w.Code, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				return fmt.Errorf("cacheable %s", check.path)
			}
			for _, s := range check.fragments {
				if !strings.Contains(w.Body.String(), s) {
					return fmt.Errorf("%s missing %s: %s", check.path, s, w.Body.String())
				}
			}
			for _, secret := range []string{"passwordHash", "sessionHash", "tokenHash", "grants", fixturePassword} {
				if strings.Contains(w.Body.String(), secret) {
					return fmt.Errorf("secret projection at %s", check.path)
				}
			}
		}
		w := perform(e, "GET", "/v1/users", "", boOrigin, "", cookies)
		userRead := perform(e, "GET", "/v1/users/"+org.User.ID.String(), "", boOrigin, "", cookies)
		var userDetail v1.V1UserGetBody
		if err = json.Unmarshal(userRead.Body.Bytes(), &userDetail); err != nil {
			return err
		}
		if userDetail.Revision != customers.UserRevision(org.User) || userDetail.CanChangeStatus || userDetail.StatusBlockedReason == "" {
			return fmt.Errorf("user revision/final-owner eligibility mismatch: %s", userRead.Body.String())
		}
		membershipRead := perform(e, "GET", "/v1/organizations/"+org.Organization.ID.String()+"/memberships/"+org.Owner.ID.String(), "", boOrigin, "", cookies)
		var memberDetail v1.V1OrganizationMembershipGetBody
		if err = json.Unmarshal(membershipRead.Body.Bytes(), &memberDetail); err != nil {
			return err
		}
		if memberDetail.Revision != customers.MembershipRevision(org.Owner) || memberDetail.CanRemove || memberDetail.CanChangeRole {
			return fmt.Errorf("owner revision/eligibility mismatch: %s", membershipRead.Body.String())
		}
		var page struct {
			Items []json.RawMessage `json:"items"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			return err
		}
		if len(page.Items) != 25 {
			return fmt.Errorf("default page length %d", len(page.Items))
		}
		if w := perform(e, "GET", "/v1/users", "", boOrigin, "", nil); w.Code != 401 {
			return fmt.Errorf("anonymous read %d", w.Code)
		}
		in := bootstrapInput("admin-read@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		adminCookies, _, err := loginHTTP(e, admin.Email)
		if err != nil {
			return err
		}
		for _, path := range []string{"/v1/users", "/v1/users/" + org.User.ID.String(), "/v1/users/choices", "/v1/organizations", "/v1/organizations/" + org.Organization.ID.String(), "/v1/organizations/" + org.Organization.ID.String() + "/memberships", "/v1/organizations/" + org.Organization.ID.String() + "/memberships/" + org.Owner.ID.String(), "/v1/users/" + org.User.ID.String() + "/memberships"} {
			w := perform(e, "GET", path, "", boOrigin, "", adminCookies)
			if w.Code != 200 {
				return fmt.Errorf("Admin read %s: %d", path, w.Code)
			}
			for _, forbidden := range []string{`"canChangeStatus":true`, `"canChangeEmail":true`, `"canVerifyEmail":true`, `"canChangeRole":true`} {
				if strings.Contains(w.Body.String(), forbidden) {
					return fmt.Errorf("Admin eligibility escalated %s", path)
				}
			}
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, member.ID); err != nil {
			return err
		}
		choice := perform(e, "GET", "/v1/users/choices?q=absent&organizationId="+org.Organization.ID.String()+"&selectedId="+member.ID.String(), "", boOrigin, "", cookies)
		var choices v1.V1UserChoiceListBody
		if err = json.Unmarshal(choice.Body.Bytes(), &choices); err != nil {
			return err
		}
		if choice.Code != 200 || len(choices.Items) != 0 || choices.Selected == nil || choices.Selected.Eligible || !strings.Contains(choices.Selected.DisabledReason, "pemulihan") {
			return fmt.Errorf("historical active-user selection lost restore eligibility: %s", choice.Body.String())
		}
		return nil
	})
}

// Instrument the real pgx read only to coordinate a committed concurrent writer
// after the item query. No SQL, authorization, or database result is substituted.
type customerReadBarrier struct {
	once             sync.Once
	reached, release chan struct{}
}

func (b *customerReadBarrier) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	if strings.HasPrefix(d.SQL, "-- name: BOListCustomerUsers :many") {
		return context.WithValue(ctx, customerReadBarrierKey{}, true)
	}
	return ctx
}

type customerReadBarrierKey struct{}

func (b *customerReadBarrier) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	if ctx.Value(customerReadBarrierKey{}) == true {
		b.once.Do(func() {
			close(b.reached)
			select {
			case <-b.release:
			case <-ctx.Done():
			}
		})
	}
}
func TestBOCustomerReadSnapshotConsistency(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "snapshot@example.test", DisplayName: "Snapshot"})
		if err != nil {
			return err
		}
		barrier := &customerReadBarrier{reached: make(chan struct{}), release: make(chan struct{})}
		cfg := p.Config()
		cfg.ConnConfig.Tracer = barrier
		reads, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			return err
		}
		defer reads.Close()
		e, _ := customerRouter(reads, svc, func() time.Time { return fixtureNow })
		cookies, _, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		done := make(chan error, 1)
		go func() {
			w := perform(e, "GET", "/v1/users?status=active", "", boOrigin, "", cookies)
			if w.Code != 200 {
				done <- fmt.Errorf("snapshot status %d", w.Code)
				return
			}
			var body v1.V1UserListBody
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				done <- err
				return
			}
			if len(body.Items) != 1 || body.Items[0].Status != "active" || body.Pagination.Total != "1" || body.Summary.Active != "1" || body.Summary.Disabled != "0" {
				done <- fmt.Errorf("mixed read snapshot: %s", w.Body.String())
				return
			}
			done <- nil
		}()
		select {
		case <-barrier.reached:
		case <-ctx.Done():
			close(barrier.release)
			<-done
			return ctx.Err()
		}
		_, err = p.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u.ID)
		close(barrier.release)
		readErr := <-done
		if err != nil {
			return err
		}
		return readErr
	})
}
