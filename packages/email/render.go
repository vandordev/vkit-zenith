// Package email renders committed templates and delivers bounded transactional
// messages. It has no account rules, database transaction or JavaScript runtime.
package email

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
	texttemplate "text/template"
	"time"
	"unicode/utf8"
)

//go:embed templates/*.tmpl templates/manifest.json
var artifacts embed.FS

type Purpose string

const (
	VerifyEmail   Purpose = "verify_email"
	SetupPassword Purpose = "setup_password"
	ResetPassword Purpose = "reset_password"
)

var ErrInvalidData = errors.New("email: invalid template data")
var ErrUnavailable = errors.New("email: template unavailable")

type TemplateData struct {
	DisplayName, ActionURL string
	ExpiresAt              time.Time
}

func (TemplateData) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted email template data]")
}
func (TemplateData) MarshalJSON() ([]byte, error) {
	return nil, errors.New("email template data is not serializable")
}

type Message struct{ Subject, HTML, Text string }

func (Message) Format(w fmt.State, _ rune) { _, _ = io.WriteString(w, "[redacted email message]") }
func (Message) MarshalJSON() ([]byte, error) {
	return nil, errors.New("email message is not serializable")
}

type manifestEntry struct{ Subject, HTML, Text string }
type manifest struct {
	Source       string
	SourceSHA256 string
	Fields       []string
	Templates    map[Purpose]manifestEntry
}
type renderedTemplate struct {
	subject string
	html    *template.Template
	text    *texttemplate.Template
}
type Renderer struct {
	origin    string
	templates map[Purpose]renderedTemplate
}

var opaque = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func NewRenderer(origin string, development bool) (*Renderer, error) {
	u, err := url.Parse(origin)
	if err != nil || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" || u.ForceQuery || u.String() != origin {
		return nil, ErrInvalidData
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if !development || u.Scheme != "http" || (u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return nil, ErrInvalidData
		}
	}
	r := &Renderer{origin: origin, templates: make(map[Purpose]renderedTemplate)}
	b, err := artifacts.ReadFile("templates/manifest.json")
	if err != nil {
		return nil, ErrUnavailable
	}
	var m manifest
	if json.Unmarshal(b, &m) != nil || len(m.Templates) != 3 || m.Source != "packages/ui-email" || strings.Join(m.Fields, ",") != "DisplayName,ActionURL,ExpiresAt" {
		return nil, ErrUnavailable
	}
	for _, purpose := range []Purpose{VerifyEmail, SetupPassword, ResetPassword} {
		entry, ok := m.Templates[purpose]
		if !ok || entry.Subject == "" || strings.ContainsAny(entry.Subject, "\r\n") || entry.HTML != string(purpose)+".html.tmpl" || entry.Text != string(purpose)+".txt.tmpl" {
			return nil, ErrUnavailable
		}
		h, err := artifacts.ReadFile("templates/" + entry.HTML)
		if err != nil {
			return nil, ErrUnavailable
		}
		t, err := artifacts.ReadFile("templates/" + entry.Text)
		if err != nil {
			return nil, ErrUnavailable
		}
		htmlTemplate, err := template.New("email").Delims("[[", "]]").Option("missingkey=error").Parse(string(h))
		if err != nil {
			return nil, ErrUnavailable
		}
		textTemplate, err := texttemplate.New("email").Delims("[[", "]]").Option("missingkey=error").Parse(string(t))
		if err != nil {
			return nil, ErrUnavailable
		}
		r.templates[purpose] = renderedTemplate{entry.Subject, htmlTemplate, textTemplate}
	}
	return r, nil
}

func actionPath(p Purpose) string {
	switch p {
	case VerifyEmail:
		return "/verify-email"
	case SetupPassword:
		return "/set-password"
	case ResetPassword:
		return "/reset-password"
	}
	return ""
}

func (r *Renderer) Render(purpose Purpose, data TemplateData) (Message, error) {
	if r == nil {
		return Message{}, ErrUnavailable
	}
	t, ok := r.templates[purpose]
	if !ok {
		return Message{}, ErrInvalidData
	}
	if !utf8.ValidString(data.DisplayName) || strings.TrimSpace(data.DisplayName) == "" || utf8.RuneCountInString(data.DisplayName) > 200 || strings.ContainsAny(data.DisplayName, "\x00\r\n") || data.ExpiresAt.IsZero() || data.ExpiresAt.Year() < 1 || data.ExpiresAt.Year() > 9999 {
		return Message{}, ErrInvalidData
	}
	u, err := url.Parse(data.ActionURL)
	if err != nil || u.User != nil || u.Fragment != "" || u.Opaque != "" || u.Scheme+"://"+u.Host != r.origin || u.Path != actionPath(purpose) || u.RawPath != "" || len(data.ActionURL) > 2048 {
		return Message{}, ErrInvalidData
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(q) < 1 || len(q) > 2 || len(q["token"]) != 1 || !opaque.MatchString(q.Get("token")) {
		return Message{}, ErrInvalidData
	}
	for key, values := range q {
		if (key != "token" && key != "request") || len(values) != 1 || !opaque.MatchString(values[0]) {
			return Message{}, ErrInvalidData
		}
	}
	if q.Encode() != u.RawQuery {
		return Message{}, ErrInvalidData
	}
	fields := struct{ DisplayName, ActionURL, ExpiresAt string }{data.DisplayName, data.ActionURL, data.ExpiresAt.UTC().Format("02 Jan 2006 15:04 UTC")}
	var h, text bytes.Buffer
	if t.html.Execute(&h, fields) != nil || t.text.Execute(&text, fields) != nil || h.Len() > 512*1024 || text.Len() > 64*1024 {
		return Message{}, ErrUnavailable
	}
	return Message{t.subject, h.String(), text.String()}, nil
}
