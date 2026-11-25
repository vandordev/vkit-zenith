package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type emailRoundTrip func(*http.Request) (*http.Response, error)

func (f emailRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func emailSendInput() SendInput {
	return SendInput{AttemptID: "fake-attempt-00000001", Recipient: "recipient@example.test", Message: Message{Subject: "Fake subject", HTML: "<p>Fake HTML</p>", Text: "Fake plain text"}}
}

func TestAuthEmailResendPayloadAndIdempotency(t *testing.T) {
	calls := 0
	transport := emailRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) <= 0 {
			t.Error("missing bounded send deadline")
		}
		if r.Method != "POST" || r.URL.String() != "https://api.resend.com/emails" || r.Header.Get("Authorization") != "Bearer fake-resend-key" || r.Header.Get("Idempotency-Key") != "fake-attempt-00000001" || r.GetBody != nil {
			t.Error("unsafe provider request or replayable body")
		}
		var body struct {
			From                string
			To                  []string
			Subject, HTML, Text string
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.From != "Zenith <noreply@example.test>" || len(body.To) != 1 || body.To[0] != "recipient@example.test" || body.HTML != "<p>Fake HTML</p>" || body.Text != "Fake plain text" {
			t.Error("SDK payload differs from rendered message")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"fake-provider-id"}`))}, nil
	})
	s, err := NewResend("fake-resend-key", "Zenith <noreply@example.test>", transport)
	if err != nil {
		t.Fatal("sender construction failed")
	}
	if err = s.Send(context.Background(), emailSendInput()); err != nil {
		t.Fatal("accepted send failed")
	}
	if calls != 1 {
		t.Fatal("delivery automatically retried")
	}
}

func TestAuthEmailResendSafeFailures(t *testing.T) {
	for _, mode := range []string{"reject", "timeout", "network", "redirect", "malformed", "oversize"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			transport := emailRoundTrip(func(r *http.Request) (*http.Response, error) {
				calls++
				if mode == "timeout" {
					<-r.Context().Done()
					return nil, r.Context().Err()
				}
				if mode == "network" {
					return nil, errors.New("sensitive fake recipient/token provider error")
				}
				status := 400
				body := `{"message":"sensitive fake recipient/token provider error"}`
				if mode == "redirect" {
					status = 307
				}
				if mode == "malformed" {
					status = 200
					body = `invalid sensitive fake response`
				}
				if mode == "oversize" {
					status = 200
					body = `{"id":"` + strings.Repeat("X", 100000) + `"}`
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"https://evil.example.test"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			s, err := NewResend("fake-resend-key", "noreply@example.test", transport)
			if err != nil {
				t.Fatal("sender construction failed")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			err = s.Send(ctx, emailSendInput())
			if err == nil || strings.Contains(err.Error(), "sensitive") || strings.Contains(err.Error(), "recipient") || strings.Contains(err.Error(), "token") {
				t.Fatal("provider failure missing or not sanitized")
			}
			if mode == "reject" || mode == "redirect" {
				if !errors.Is(err, ErrRejected) {
					t.Fatal("explicit rejection not classified")
				}
			} else if !errors.Is(err, ErrUnknown) {
				t.Fatal("uncertain delivery reported as known failure")
			}
			if calls != 1 {
				t.Fatal("failure automatically retried")
			}
		})
	}
}

func TestAuthEmailResendInvalidAndCanceledDoNotSend(t *testing.T) {
	calls := 0
	s, err := NewResend("fake-resend-key", "noreply@example.test", emailRoundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected provider call")
	}))
	if err != nil {
		t.Fatal("sender construction failed")
	}
	input := emailSendInput()
	input.Recipient = "bad\r\nrecipient"
	if err = s.Send(context.Background(), input); !errors.Is(err, ErrNotSent) {
		t.Fatal("invalid recipient sent")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err = s.Send(ctx, emailSendInput()); !errors.Is(err, ErrNotSent) {
		t.Fatal("pre-canceled send attempted")
	}
	if calls != 0 {
		t.Fatal("invalid request reached provider")
	}
}

func TestAuthEmailDebugValuesCannotEscape(t *testing.T) {
	s, err := NewResend("sensitive-fake-key", "noreply@example.test", emailRoundTrip(func(*http.Request) (*http.Response, error) { return nil, ErrNotSent }))
	if err != nil {
		t.Fatal("sender construction failed")
	}
	for _, value := range []any{s, emailSendInput(), emailTestData(VerifyEmail)} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
			out := fmt.Sprintf(format, value)
			if strings.Contains(out, "sensitive-fake-key") || strings.Contains(out, "recipient@example.test") || strings.Contains(out, "auth.example.test") {
				t.Fatal("ephemeral email boundary leaked through debug formatting")
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatal("ephemeral email boundary generically serialized")
		}
	}
}
