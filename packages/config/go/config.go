// Package config implements the shared, consumer-isolated YAML contract.
package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Options struct {
	Root string
	// Environment is a complete snapshot when non-nil, not an overlay.
	Environment map[string]string
}
type AppConfig struct {
	Environment string `validate:"oneof=development staging production test"`
	LogLevel    string `validate:"oneof=debug info warn error"`
}
type BasicAuthConfig struct {
	Enabled  bool
	Username string `validate:"required_if=Enabled true"`
	Password string `validate:"required_if=Enabled true"`
}
type DocsConfig struct{ BasicAuth BasicAuthConfig }
type WebOrigin string

type ServiceConfig struct {
	Host                   string `validate:"required"`
	Port                   int    `validate:"min=1,max=65535"`
	ReadTimeoutSeconds     int    `validate:"min=1,max=3600"`
	WriteTimeoutSeconds    int    `validate:"min=1,max=3600"`
	IdleTimeoutSeconds     int    `validate:"min=1,max=3600"`
	ShutdownTimeoutSeconds int    `validate:"min=1,max=3600"`
	Docs                   DocsConfig
}
type APIConfig struct {
	App AppConfig
	API ServiceConfig
}

func LoadAPI(options Options) (APIConfig, error) {
	d, err := loadDocument(options, "api")
	if err != nil {
		return APIConfig{}, err
	}
	app, service, err := d.selected("api")
	return APIConfig{App: app, API: service}, err
}

type WebAPIConfig = APIConfig

func LoadWebAPI(options Options) (WebAPIConfig, error) {
	d, err := loadDocument(options, "webApi")
	if err != nil {
		return WebAPIConfig{}, err
	}
	app, service, err := d.selected("webApi")
	return WebAPIConfig{App: app, API: service}, err
}
func validOrigin(value string) bool {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || !strings.HasPrefix(value, u.Scheme+"://") || u.Host == "" || u.User != nil || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Path != "" || u.RawPath != "" || strings.ContainsAny(value, "# \t\r\n") || u.Host != strings.ToLower(u.Host) || strings.HasSuffix(u.Host, ":") {
		return false
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.String() != host {
			return false
		}
	} else {
		if host == "" || len(host) > 253 || strings.HasSuffix(host, ".") || strings.Trim(host, "0123456789.") == "" {
			return false
		}
		labels := strings.Split(host, ".")
		for _, label := range labels {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return false
			}
			for _, r := range label {
				if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
					return false
				}
			}
		}
		last := labels[len(labels)-1]
		if strings.Trim(last, "0123456789") == "" || (strings.HasPrefix(last, "0x") && strings.Trim(last[2:], "0123456789abcdef") == "") {
			return false
		}
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port || (u.Scheme == "http" && n == 80) || (u.Scheme == "https" && n == 443) {
			return false
		}
	}
	return true
}
func (d *document) validate(value any) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(value); err != nil {
		if fields, ok := err.(validator.ValidationErrors); ok && len(fields) > 0 {
			return d.fail(fields[0].StructNamespace(), "invalid value or incomplete credentials")
		}
		return d.fail("document", "invalid typed configuration")
	}
	return nil
}
func (d *document) selected(name string) (AppConfig, ServiceConfig, error) {
	return d.selectedService(name)
}

func (d *document) selectedService(name string) (AppConfig, ServiceConfig, error) {
	var app AppConfig
	var service ServiceConfig
	var err error
	if app.Environment, err = d.stringAt("app", "environment"); err != nil {
		return app, service, err
	}
	if app.LogLevel, err = d.stringAt("app", "logLevel"); err != nil {
		return app, service, err
	}
	if service.Host, err = d.stringAt(name, "host"); err != nil {
		return app, service, err
	}
	for _, target := range []struct {
		key string
		out *int
	}{
		{"port", &service.Port}, {"readTimeoutSeconds", &service.ReadTimeoutSeconds}, {"writeTimeoutSeconds", &service.WriteTimeoutSeconds}, {"idleTimeoutSeconds", &service.IdleTimeoutSeconds}, {"shutdownTimeoutSeconds", &service.ShutdownTimeoutSeconds},
	} {
		value, err := d.valueAt(name, target.key)
		if err != nil {
			return app, service, err
		}
		switch n := value.(type) {
		case string:
			if !digits.MatchString(n) {
				return app, service, d.fail(name+"."+target.key, "expected integer")
			}
			*target.out, err = strconv.Atoi(n)
		case float64:
			if n != float64(int(n)) {
				err = fmt.Errorf("not integer")
			} else {
				*target.out = int(n)
			}
		default:
			err = fmt.Errorf("not numeric")
		}
		if err != nil {
			return app, service, d.fail(name+"."+target.key, "expected integer")
		}
	}
	enabled, err := d.valueAt(name, "docs", "basicAuth", "enabled")
	if err != nil {
		return app, service, err
	}
	switch value := enabled.(type) {
	case bool:
		service.Docs.BasicAuth.Enabled = value
	case string:
		if value != "true" && value != "false" {
			return app, service, d.fail(name+".docs.basicAuth.enabled", "expected boolean")
		}
		service.Docs.BasicAuth.Enabled = value == "true"
	default:
		return app, service, d.fail(name+".docs.basicAuth.enabled", "expected boolean")
	}
	// Disabled credentials are never selected or interpolated. Raw grammar is
	// still checked globally, but unrelated missing secrets cannot block startup.
	if service.Docs.BasicAuth.Enabled {
		if service.Docs.BasicAuth.Username, err = d.stringAt(name, "docs", "basicAuth", "username"); err != nil {
			return app, service, err
		}
		if service.Docs.BasicAuth.Password, err = d.stringAt(name, "docs", "basicAuth", "password"); err != nil {
			return app, service, err
		}
	}
	if err := d.validate(struct {
		App     AppConfig
		Service ServiceConfig
	}{app, service}); err != nil {
		return app, service, err
	}
	return app, service, nil
}
