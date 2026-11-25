package authtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// EmailCapture is an injected test-only SDK transport, not an HTTP listener or
// production configuration option. It only captures the real rendered payload.
type EmailCapture struct {
	mu       sync.Mutex
	capacity int
	messages []CapturedEmail
}
type CapturedEmail struct {
	AttemptID, From     string
	To                  []string
	Subject, HTML, Text string
}

func (CapturedEmail) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted captured email]")
}
func (CapturedEmail) MarshalJSON() ([]byte, error) {
	return nil, errors.New("captured email is not serializable")
}
func (*EmailCapture) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted email capture]")
}

func NewEmailCapture(capacity int) *EmailCapture {
	if capacity < 1 || capacity > 1000 {
		panic("test email capture: capacity must be bounded")
	}
	return &EmailCapture{capacity: capacity}
}

func (c *EmailCapture) RoundTrip(r *http.Request) (*http.Response, error) {
	if r == nil || r.URL == nil || r.Method != "POST" || r.URL.String() != "https://api.resend.com/emails" || r.Body == nil {
		return nil, errors.New("test email capture: unexpected destination")
	}
	if r.Context().Err() != nil {
		return nil, errors.New("test email capture: canceled")
	}
	defer r.Body.Close()
	var data struct {
		From                string
		To                  []string
		Subject, HTML, Text string
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1024*1024))
	if decoder.Decode(&data) != nil || len(data.To) != 1 || data.From == "" || data.HTML == "" || data.Text == "" || data.Subject == "" || r.Header.Get("Idempotency-Key") == "" {
		return nil, errors.New("test email capture: invalid rendered message")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, errors.New("test email capture: invalid payload boundary")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.messages) >= c.capacity {
		return nil, errors.New("test email capture: full")
	}
	c.messages = append(c.messages, CapturedEmail{AttemptID: r.Header.Get("Idempotency-Key"), From: data.From, To: append([]string(nil), data.To...), Subject: data.Subject, HTML: data.HTML, Text: data.Text})
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"owned-test-capture"}`)), Request: r}, nil
}

func (c *EmailCapture) Snapshot() []CapturedEmail {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]CapturedEmail, len(c.messages))
	for i, item := range c.messages {
		out[i] = item
		out[i].To = append([]string(nil), item.To...)
	}
	return out
}
func (c *EmailCapture) Clear() { c.mu.Lock(); defer c.mu.Unlock(); clear(c.messages); c.messages = nil }
