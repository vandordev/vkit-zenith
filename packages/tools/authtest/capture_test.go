//go:build integration

package authtest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/email"
)

type emailRoundTrip func(*http.Request) (*http.Response, error)

func (f emailRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAuthEmailCaptureBoundary(t *testing.T) {
	c := NewEmailCapture(1)
	renderer, err := email.NewRenderer("https://auth.example.test", false)
	if err != nil {
		t.Fatal("renderer unavailable")
	}
	message, err := renderer.Render(email.VerifyEmail, email.TemplateData{DisplayName: "Rani <Pengguna>", ActionURL: "https://auth.example.test/verify-email?token=" + strings.Repeat("A", 43), ExpiresAt: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal("actual email template render failed")
	}
	request := func(key string) *http.Request {
		body, err := json.Marshal(struct {
			From    string   `json:"from"`
			To      []string `json:"to"`
			Subject string   `json:"subject"`
			HTML    string   `json:"html"`
			Text    string   `json:"text"`
		}{"noreply@example.test", []string{"recipient@example.test"}, message.Subject, message.HTML, message.Text})
		if err != nil {
			t.Fatal("rendered message encoding failed")
		}
		r, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
		if err != nil {
			t.Fatal("request unavailable")
		}
		r.Header.Set("Idempotency-Key", key)
		return r
	}
	response, err := c.RoundTrip(request("fake-attempt-00000001"))
	if err != nil {
		t.Fatal("capture failed")
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	items := c.Snapshot()
	if len(items) != 1 || items[0].HTML != message.HTML || items[0].Text != message.Text || items[0].Subject != message.Subject || items[0].AttemptID != "fake-attempt-00000001" {
		t.Fatal("capture did not preserve actual rendered email")
	}
	items[0].To[0] = "mutated@example.test"
	if c.Snapshot()[0].To[0] != "recipient@example.test" {
		t.Fatal("capture snapshots alias protected inbox")
	}
	if _, err = c.RoundTrip(request("fake-attempt-00000002")); err == nil {
		t.Fatal("capture capacity unbounded")
	}
	if strings.Contains(fmt.Sprintf("%+v", c.Snapshot()), "recipient@example.test") {
		t.Fatal("capture debug leaked recipient")
	}
	if _, err = json.Marshal(c.Snapshot()); err == nil {
		t.Fatal("capture generically serialized")
	}
	for _, r := range []*http.Request{request("missing"), request("fake-attempt-00000003")} {
		r.URL.Host = "unexpected.example.test"
		if _, err := NewEmailCapture(1).RoundTrip(r); err == nil {
			t.Fatal("capture silently accepted unexpected network destination")
		}
	}
	c.Clear()
	if len(c.Snapshot()) != 0 {
		t.Fatal("capture clear failed")
	}
}

func TestAuthEmailDeliveryPostCommit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
		if err := db.InTx(ctx, f.Pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
			for _, query := range []string{
				`INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000001','recipient@example.test','recipient@example.test','Recipient','active',$1,$1,1)`,
				`INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES(gen_random_uuid(),'10000000-0000-4000-8000-000000000001','setup_password','recipient@example.test',sha256('fake-token'::bytea),$1,$1::timestamptz+interval '24 hours',1,0)`,
				`INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,detail) VALUES(gen_random_uuid(),$1,'system','identity.fixture','success','{}')`,
			} {
				if _, err := tx.Exec(ctx, query, at); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		capture := NewEmailCapture(1)
		renderer, err := email.NewRenderer("https://auth.example.test", false)
		if err != nil {
			return err
		}
		token := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))
		message, err := renderer.Render(email.SetupPassword, email.TemplateData{DisplayName: "Recipient", ActionURL: "https://auth.example.test/set-password?token=" + token, ExpiresAt: at.Add(24 * time.Hour)})
		if err != nil {
			return email.ErrNotSent
		}
		transport := emailRoundTrip(func(r *http.Request) (*http.Response, error) {
			tx, err := f.Pool.Begin(ctx)
			if err != nil {
				return nil, err
			}
			defer tx.Rollback(ctx)
			var active bool
			if err = tx.QueryRow(ctx, `SELECT status='active' FROM users WHERE id='10000000-0000-4000-8000-000000000001' FOR UPDATE NOWAIT`).Scan(&active); err != nil {
				return nil, err
			}
			if !active {
				return nil, errors.New("post-commit identity state unavailable")
			}
			if err = tx.Commit(ctx); err != nil {
				return nil, err
			}
			return capture.RoundTrip(r)
		})
		sender, err := email.NewResend("fake-resend-key", "noreply@example.test", transport)
		if err != nil {
			return err
		}
		if err = sender.Send(ctx, email.SendInput{AttemptID: "fake-attempt-00000001", Recipient: "recipient@example.test", Message: message}); err != nil {
			return err
		}
		if len(capture.Snapshot()) != 1 {
			return errors.New("post-commit rendered capture missing")
		}
		var rows int
		if err = f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users)+(SELECT count(*) FROM email_action_tokens)+(SELECT count(*) FROM audit_events)`).Scan(&rows); err != nil {
			return err
		}
		if rows != 3 {
			return errors.New("post-commit records changed")
		}
		if !strings.Contains(capture.Snapshot()[0].HTML, "Buat kata sandi") || !strings.Contains(capture.Snapshot()[0].Text, token) {
			return errors.New("capture did not contain rendered password setup")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned PostgreSQL post-commit capture failed; raw details withheld")
	}
}
