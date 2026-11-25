package email

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/resend/resend-go/v4"
)

var (
	ErrNotSent  = errors.New("email: not sent")
	ErrRejected = errors.New("email: provider rejected request")
	ErrUnknown  = errors.New("email: provider acceptance unknown")
)

type SendInput struct {
	AttemptID, Recipient string
	Message              Message
}
type Transport interface {
	Send(context.Context, SendInput) error
}
type Resend struct {
	key, from string
	transport http.RoundTripper
}

func (*Resend) Format(w fmt.State, _ rune) { _, _ = io.WriteString(w, "[redacted email sender]") }
func (*Resend) MarshalJSON() ([]byte, error) {
	return nil, errors.New("email sender is not serializable")
}
func (SendInput) Format(w fmt.State, _ rune) { _, _ = io.WriteString(w, "[redacted email send input]") }
func (SendInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("email send input is not serializable")
}

var attemptID = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

func NewResend(key, from string, transport http.RoundTripper) (*Resend, error) {
	a, err := mail.ParseAddress(from)
	if err != nil || a.Address == "" || len(from) > 320 || strings.ContainsAny(from, "\r\n") || len(key) < 8 || len(key) > 4096 {
		return nil, ErrNotSent
	}
	if transport == nil {
		// Purpose-owned transport, no ambient proxy lookup or redirects carrying
		// recipient/link/credential bodies to another destination.
		transport = &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 5, MaxConnsPerHost: 10, IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 5 * time.Second, ForceAttemptHTTP2: true}
	}
	return &Resend{key, from, transport}, nil
}

type boundedProviderTransport struct {
	next   http.RoundTripper
	status int
}
type limitedBody struct {
	io.Reader
	io.Closer
}

func (t *boundedProviderTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != "POST" || r.URL.String() != "https://api.resend.com/emails" {
		return nil, ErrNotSent
	}
	r = r.Clone(r.Context())
	// net/http otherwise treats Idempotency-Key + GetBody as permission to
	// retry a POST after a reused connection fails. Default policy is one attempt.
	r.GetBody = nil
	response, err := t.next.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Body == nil {
		return nil, ErrUnknown
	}
	t.status = response.StatusCode
	response.Body = limitedBody{io.LimitReader(response.Body, 64*1024), response.Body}
	return response, nil
}

// Send performs one bounded SDK call. A nil error means provider API acceptance,
// never inbox placement/delivery. No response body or provider error escapes.
func (s *Resend) Send(ctx context.Context, input SendInput) error {
	if s == nil || ctx.Err() != nil || !attemptID.MatchString(input.AttemptID) || len(input.Recipient) > 254 || strings.ContainsAny(input.Recipient, "\r\n") {
		return ErrNotSent
	}
	a, err := mail.ParseAddress(input.Recipient)
	if err != nil || a.Address != input.Recipient || input.Message.Subject == "" || strings.ContainsAny(input.Message.Subject, "\r\n") || input.Message.HTML == "" || input.Message.Text == "" || len(input.Message.HTML) > 512*1024 || len(input.Message.Text) > 64*1024 {
		return ErrNotSent
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	t := &boundedProviderTransport{next: s.transport}
	httpClient := &http.Client{Transport: t, Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	client := resend.NewCustomClient(httpClient, s.key)
	// The pinned SDK reads RESEND_BASE_URL at package initialization. Its value
	// is never an authority here: explicitly replace it with the fixed endpoint.
	client.BaseURL = &url.URL{Scheme: "https", Host: "api.resend.com", Path: "/"}
	result, err := client.Emails.SendWithOptions(ctx, &resend.SendEmailRequest{From: s.from, To: []string{input.Recipient}, Subject: input.Message.Subject, Html: input.Message.HTML, Text: input.Message.Text}, &resend.SendEmailOptions{IdempotencyKey: input.AttemptID})
	if err != nil {
		if t.status >= 300 && t.status < 500 {
			return ErrRejected
		}
		return ErrUnknown
	}
	if result == nil || result.Id == "" || len(result.Id) > 128 {
		return ErrUnknown
	}
	return nil
}
