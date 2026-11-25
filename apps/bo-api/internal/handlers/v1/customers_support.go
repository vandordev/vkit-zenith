package v1

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

type CustomerPage struct {
	Page     int    `query:"page" default:"1" minimum:"1" maximum:"1000000"`
	PageSize int    `query:"pageSize" default:"25" minimum:"1" maximum:"100"`
	Q        string `query:"q" maxLength:"100"`
}
type customerUserPage struct {
	CustomerPage
	Status       string `query:"status" enum:",active,disabled"`
	Verification string `query:"verification" enum:",verified,unverified"`
}
type customerOrganizationPage struct {
	CustomerPage
	Status string `query:"status" enum:",active,suspended"`
}
type CustomerMembershipPage struct {
	CustomerPage
	Status    string `query:"status" enum:",active,removed"`
	Ownership string `query:"ownership" enum:",owner,member"`
}
type CustomerOrganizationPath struct {
	OrganizationID string `path:"organizationId" format:"uuid"`
}
type CustomerMembershipPath struct {
	CustomerOrganizationPath
	ID string `path:"id" format:"uuid"`
}

type CustomerSensitiveInput struct {
	Revision        string `json:"revision" minLength:"64" maxLength:"64" pattern:"^[a-fA-F0-9]{64}$"`
	Reason          string `json:"reason" minLength:"1" maxLength:"1024"`
	CurrentPassword string `json:"currentPassword" minLength:"1" maxLength:"512"`
}

func (s CustomerSensitiveInput) domain() customers.Sensitive {
	return customers.Sensitive{Revision: s.Revision, Reason: s.Reason, CurrentPassword: s.CurrentPassword}
}

func customerUUID(s string) pgtype.UUID { var id pgtype.UUID; _ = id.Scan(s); return id }
func customerCount(n int64) string      { return strconv.FormatInt(n, 10) }
func customerPattern(s string) string {
	return "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s) + "%"
}
func customerTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time.UTC()
	return &v
}
func customerError(ctx context.Context, err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, staff.ErrNotFound) {
		return &huma.ErrorModel{Type: "urn:zenith:bo:not_found", Title: "Not Found", Status: 404, Detail: "Data tidak tersedia."}
	}
	return security.Error(ctx, err)
}
func customerSnapshot(ctx context.Context, b *security.Boundary) (*persistence.Queries, pgx.Tx, staff.Staff, error) {
	actor, err := b.Read(ctx, "any_staff")
	if err != nil {
		return nil, nil, staff.Staff{}, customerError(ctx, err)
	}
	tx, err := b.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, staff.Staff{}, customerError(ctx, err)
	}
	return persistence.New(tx), tx, actor, nil
}
func customerCommands(b *security.Boundary) *customers.Service {
	return customers.NewService(b.Pool, b.Clock, b.Service)
}
func customerUserDomain(r persistence.BOListCustomerUsersRow) customers.User {
	return customers.User{ID: r.ID, Email: r.Email, DisplayName: r.DisplayName, Status: r.Status, EmailVerifiedAt: customerTime(r.EmailVerifiedAt), CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time}
}
func customerOrganizationDomain(r persistence.BOListCustomerOrganizationsRow) customers.Organization {
	return customers.Organization{ID: r.ID, Name: r.Name, Slug: r.Slug, Status: r.Status, CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time}
}
func customerMembershipDomain(r persistence.BOListCustomerMembershipsRow) customers.Membership {
	return customers.Membership{ID: r.ID, OrganizationID: r.OrganizationID, UserID: r.UserID, Status: r.Status, Role: customers.Role(r.Role), CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time}
}

func customerSensitiveReason(super bool) string {
	if !super {
		return "Hanya Superadmin dengan konfirmasi kata sandi."
	}
	return ""
}
func customerDisableReason(super, canDisable bool, status string) string {
	if !super {
		return customerSensitiveReason(false)
	}
	if status == "active" && !canDisable {
		return "Pengguna adalah pemilik aktif terakhir. Tambahkan pemilik pengganti di organisasi terkait terlebih dahulu."
	}
	return ""
}
func customerVerifyReason(super bool, verified bool) string {
	if !super {
		return customerSensitiveReason(false)
	}
	if verified {
		return "Email sudah terverifikasi."
	}
	return ""
}
func customerMemberReasons(r persistence.BOListCustomerMembershipsRow, super bool) (remove, restore, owner string) {
	if r.Status != "active" {
		remove = "Keanggotaan sudah dihapus."
	} else if r.Role == "owner" {
		remove = "Cabut kepemilikan terlebih dahulu melalui tindakan terpisah."
	}
	if r.Status != "removed" {
		restore = "Keanggotaan masih aktif."
	} else if r.UserStatus != "active" {
		restore = "Pengguna dinonaktifkan."
	}
	if !super {
		owner = customerSensitiveReason(false)
	} else if r.Status != "active" {
		owner = "Kepemilikan membutuhkan keanggotaan aktif."
	} else if r.Role != "owner" && r.UserStatus != "active" {
		owner = "Menambahkan kepemilikan membutuhkan pengguna aktif."
	} else if r.Role == "owner" && !r.HasOtherOwner {
		owner = "Pemilik aktif terakhir tidak dapat dicabut."
	}
	return
}
