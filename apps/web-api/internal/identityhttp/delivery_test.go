package identityhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type deliveryTestRoundTrip func(*http.Request) (*http.Response, error)

func (f deliveryTestRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAuthDeliveryAdapterUsesRenderedPurposeBoundMessage(t *testing.T) {
	calls := 0
	sender, err := NewDeliverySender("https://auth.example.test", false, "fake-resend-key", "noreply@example.test", deliveryTestRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			Subject, Html, Text string
			To                  []string
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, err
		}
		if body.Subject != "Buat kata sandi Zenith" {
			t.Error("email subject mismatch")
		}
		if len(body.To) != 1 || body.To[0] != "rani@example.test" {
			t.Error("recipient mismatch")
		}
		if !strings.Contains(body.Html, "Siapkan kata sandi pertama Anda") {
			t.Error("rendered purpose copy missing")
		}
		if !strings.Contains(body.Html, "Rani &lt;Pemilik&gt;") {
			t.Error("HTML display-name escaping missing")
		}
		if !strings.Contains(body.Text, "Halo Rani <Pemilik>") {
			t.Error("plain-text display name missing")
		}
		if !strings.Contains(body.Text, "30 Oct 2026 12:00 UTC") {
			t.Error("expiry field missing")
		}
		if strings.Contains(body.Html, "[[.") {
			t.Error("unresolved template action")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"fake"}`))}, nil
	}))
	if err != nil {
		t.Fatal("delivery adapter setup failed")
	}
	token := strings.Repeat("A", 43)
	err = sender.Send(context.Background(), identity.ActionDelivery{AttemptID: "fake-attempt-00000001", Recipient: "rani@example.test", DisplayName: "Rani <Pemilik>", URL: "https://auth.example.test/set-password?token=" + token, Purpose: identity.ActionSetupPassword, ExpiresAt: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)})
	if err != nil || calls != 1 {
		t.Fatal("rendered delivery did not complete once")
	}
	calls = 0
	err = sender.Send(context.Background(), identity.ActionDelivery{AttemptID: "fake-attempt-00000002", Recipient: "rani@example.test", DisplayName: "Rani", URL: "https://evil.example.test/verify-email?token=" + token, Purpose: identity.ActionVerifyEmail, ExpiresAt: time.Now()})
	if err == nil || calls != 0 {
		t.Fatal("unsafe Auth URL reached provider")
	}
}
