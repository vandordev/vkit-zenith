package email

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func emailTestData(p Purpose) TemplateData {
	path := map[Purpose]string{VerifyEmail: "verify-email", SetupPassword: "set-password", ResetPassword: "reset-password"}[p]
	q := url.Values{"token": {base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))}, "request": {base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("b", 32)))}}
	return TemplateData{DisplayName: `"><img src=x onerror=alert(1)> & 世界`, ActionURL: "https://auth.example.test/" + path + "?" + q.Encode(), ExpiresAt: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)}
}

func TestAuthEmailRenderEscapesDynamicFields(t *testing.T) {
	r, err := NewRenderer("https://auth.example.test", false)
	if err != nil {
		t.Fatal("renderer construction failed")
	}
	for _, purpose := range []Purpose{VerifyEmail, SetupPassword, ResetPassword} {
		t.Run(string(purpose), func(t *testing.T) {
			data := emailTestData(purpose)
			m, err := r.Render(purpose, data)
			if err != nil {
				t.Fatal("valid rendering failed")
			}
			if strings.Contains(m.HTML, "<img src=x") || !strings.Contains(m.HTML, "&lt;img") || strings.Contains(m.HTML, "[[.") || strings.Contains(m.Text, "[[.") {
				t.Fatal("escaping or template action resolution failed")
			}
			if !strings.Contains(m.HTML, "&amp;") || !strings.Contains(m.Text, data.ActionURL) || !strings.Contains(m.Text, data.DisplayName) || !strings.Contains(m.Text, "30 Oct 2026 12:00 UTC") {
				t.Fatal("HTML/plain-text content incomplete")
			}
			links := regexp.MustCompile(`href="([^"]+)"`).FindAllStringSubmatch(m.HTML, -1)
			if len(links) != 2 {
				t.Fatal("expected action and fallback links")
			}
			for _, link := range links {
				if html.UnescapeString(link[1]) != data.ActionURL {
					t.Fatal("rendered link changed destination")
				}
			}
			if strings.Contains(fmt.Sprintf("%+v %s", m, m), "auth.example.test") {
				t.Fatal("message debug formatting leaked action URL")
			}
			if _, err = json.Marshal(m); err == nil {
				t.Fatal("rendered message escaped through generic JSON")
			}
		})
	}
}

func TestAuthEmailRenderRejectsUnsafeInputs(t *testing.T) {
	r, err := NewRenderer("https://auth.example.test", false)
	if err != nil {
		t.Fatal("renderer construction failed")
	}
	valid := emailTestData(VerifyEmail)
	for _, bad := range []string{"javascript:alert(1)", "data:text/html,test", "https://evil.example.test/verify-email?token=abc", "https://auth.example.test@evil.example.test/verify-email?token=abc", "https://auth.example.test/reset-password?token=abc", "https://auth.example.test/verify-email", "https://auth.example.test/verify-email?token=a&token=b", valid.ActionURL + "#fragment", valid.ActionURL + "&return=https://evil.example.test", strings.ReplaceAll(valid.ActionURL, "https:", "http:")} {
		data := valid
		data.ActionURL = bad
		if _, err := r.Render(VerifyEmail, data); err == nil {
			t.Fatal("unsafe URL accepted")
		}
	}
	for _, data := range []TemplateData{{}, {DisplayName: " ", ActionURL: valid.ActionURL, ExpiresAt: valid.ExpiresAt}, {DisplayName: "Name", ActionURL: valid.ActionURL}, {DisplayName: strings.Repeat("界", 201), ActionURL: valid.ActionURL, ExpiresAt: valid.ExpiresAt}} {
		if _, err := r.Render(VerifyEmail, data); err == nil {
			t.Fatal("missing or invalid fields accepted")
		}
	}
	if _, err := r.Render(Purpose("marketing"), valid); err == nil {
		t.Fatal("unapproved email purpose accepted")
	}
	for _, origin := range []string{"", "http://auth.example.test", "https://auth.example.test/path", "https://auth.example.test?x=1", "https://auth.example.test@evil.example.test", "https://auth.example.test/#x"} {
		if _, err := NewRenderer(origin, false); err == nil {
			t.Fatal("unsafe origin accepted")
		}
	}
	if _, err := NewRenderer("http://127.0.0.1:4100", true); err != nil {
		t.Fatal("explicit loopback development origin rejected")
	}
	if _, err := NewRenderer("http://evil.example.test", true); err == nil {
		t.Fatal("development bypass permits non-loopback origin")
	}
}
