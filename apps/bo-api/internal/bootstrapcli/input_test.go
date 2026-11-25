package bootstrapcli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type fixtureTerminal struct {
	tty    bool
	values []string
	err    error
}

func (f *fixtureTerminal) IsTerminal() bool { return f.tty }
func (f *fixtureTerminal) Read(_ string, _ bool) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if len(f.values) == 0 {
		return "", errors.New("cancelled")
	}
	v := f.values[0]
	f.values = f.values[1:]
	return v, nil
}
func TestBOBootstrapInput(t *testing.T) {
	for _, c := range []struct {
		name     string
		terminal fixtureTerminal
		ok       bool
	}{{"valid", fixtureTerminal{tty: true, values: []string{"root@example.test", "Fixture", "fixture password 123", "fixture password 123"}}, true}, {"mismatch", fixtureTerminal{tty: true, values: []string{"root@example.test", "Fixture", "fixture password 123", "other password 456"}}, false}, {"no_tty", fixtureTerminal{}, false}, {"cancel", fixtureTerminal{tty: true, err: context.Canceled}, false}} {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			in, err := Collect(&c.terminal, &out)
			if (err == nil) != c.ok {
				t.Fatal("input result mismatch")
			}
			if c.ok && (in.Role != "superadmin" || in.Status != "active" || in.Password != "fixture password 123") {
				t.Fatal("invalid trusted bootstrap input")
			}
			if strings.Contains(out.String(), "password 123") {
				t.Fatal("secret echoed")
			}
		})
	}
}
