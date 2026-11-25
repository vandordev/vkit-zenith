package bootstrapcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"golang.org/x/term"
)

type Terminal interface {
	IsTerminal() bool
	Read(prompt string, secret bool) (string, error)
}
type Console struct {
	In  *os.File
	Out io.Writer
}

func (c Console) IsTerminal() bool { return term.IsTerminal(int(c.In.Fd())) }
func (c Console) Read(prompt string, secret bool) (string, error) {
	if !c.IsTerminal() {
		return "", errors.New("terminal required")
	}
	if _, err := fmt.Fprint(c.Out, prompt); err != nil {
		return "", err
	}
	if secret {
		b, err := term.ReadPassword(int(c.In.Fd()))
		_, _ = fmt.Fprintln(c.Out)
		if err != nil {
			return "", errors.New("input cancelled")
		}
		value := string(b)
		for i := range b {
			b[i] = 0
		}
		return value, nil
	}
	// Bounded byte-wise read avoids buffered readers consuming subsequent secret
	// input or silently accepting arbitrarily long nonsecret lines.
	line := make([]byte, 0, 256)
	buf := make([]byte, 1)
	for len(line) <= 1024 {
		n, err := c.In.Read(buf)
		if err != nil || n != 1 {
			return "", errors.New("input cancelled")
		}
		if buf[0] == '\n' {
			return strings.TrimSuffix(string(line), "\r"), nil
		}
		line = append(line, buf[0])
	}
	return "", errors.New("input too long")
}
func Collect(t Terminal, out io.Writer) (staff.StaffWrite, error) {
	fail := func() (staff.StaffWrite, error) {
		return staff.StaffWrite{}, errors.New("Input bootstrap tidak valid atau dibatalkan.")
	}
	if !t.IsTerminal() {
		return fail()
	}
	email, err := t.Read("Email: ", false)
	if err != nil {
		return fail()
	}
	name, err := t.Read("Nama: ", false)
	if err != nil {
		return fail()
	}
	password, err := t.Read("Kata sandi: ", true)
	if err != nil {
		return fail()
	}
	confirm, err := t.Read("Ulangi kata sandi: ", true)
	if err != nil {
		return fail()
	}
	if len(email) > 254 || len(name) > 800 || !utf8.ValidString(password) || len(password) > 512 || utf8.RuneCountInString(password) < 15 || utf8.RuneCountInString(password) > 128 || password != confirm {
		return fail()
	}
	_, err = fmt.Fprintln(out, "Input diterima; membuat Superadmin pertama.")
	if err != nil {
		return fail()
	}
	return staff.StaffWrite{Email: email, DisplayName: name, Password: password, Role: "superadmin", Status: "active"}, nil
}
