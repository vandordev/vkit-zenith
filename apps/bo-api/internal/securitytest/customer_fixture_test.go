//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func customerRouter(p *pgxpool.Pool, s *staff.Service, clock func() time.Time) (*echo.Echo, *security.Boundary) {
	return httpRouterLogger(p, s, clock, slog.New(slog.NewJSONHandler(io.Discard, nil)),
		v1.NewListUsers, v1.NewGetUser, v1.NewListUserMemberships, v1.NewListUserChoices,
		v1.NewListOrganizations, v1.NewGetOrganization, v1.NewListOrganizationMemberships, v1.NewGetOrganizationMembership,
		v1.NewCreateUser, v1.NewUpdateUser, v1.NewSetUserStatus, v1.NewChangeUserEmail, v1.NewVerifyUserEmail,
		v1.NewCreateOrganization, v1.NewUpdateOrganization, v1.NewSetOrganizationStatus,
		v1.NewAddOrganizationMember, v1.NewRemoveOrganizationMember, v1.NewRestoreOrganizationMember, v1.NewSetOrganizationMembershipRole,
		v1.NewListActivity, v1.NewGetActivity, v1.NewGetRecentActivity, v1.NewListActivityActions, v1.NewListActivityStaffChoices)
}

func TestBOCustomerFixtureIsolation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("customer-root@example.test"))
		if err != nil {
			return err
		}
		e, _ := httpRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		if csrf == "" {
			return errors.New("missing fixture CSRF")
		}
		w := perform(e, "GET", "/v1/auth/me", "", boOrigin, "", cookies)
		if w.Code != 200 {
			return fmt.Errorf("fixture session status: %d", w.Code)
		}
		var n int
		if err := p.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("fixture did not begin with empty customer data")
		}
		return nil
	})
}
