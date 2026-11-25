package identityhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/vandordev/vkit-zenith/packages/application/identity"
	"github.com/vandordev/vkit-zenith/packages/email"
)

type sourceContextKey struct{}
type browserBindingKey struct{}

type BrowserBindings struct {
	Session, Enrollment identity.SecretHash
	SessionSecret       string
}

func (BrowserBindings) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted browser bindings]")
}
func (BrowserBindings) MarshalJSON() ([]byte, error) {
	return nil, errors.New("browser bindings are not serializable")
}

func WithRequestSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sourceContextKey{}, source)
}
func RequestSource(ctx context.Context) string {
	value, _ := ctx.Value(sourceContextKey{}).(string)
	return value
}
func WithBrowserBindings(ctx context.Context, b BrowserBindings) context.Context {
	return context.WithValue(ctx, browserBindingKey{}, b)
}
func RequestBrowserBindings(ctx context.Context) BrowserBindings {
	b, _ := ctx.Value(browserBindingKey{}).(BrowserBindings)
	return b
}

// ResolveSource honors X-Forwarded-For only when the immediate network peer is
// inside an explicitly configured trusted-proxy CIDR. It walks trusted hops from
// right to left and chooses the first untrusted address; leftmost attacker input
// can never override the network-adjacent client identity.
func ResolveSource(remoteAddr, forwardedFor string, trusted []netip.Prefix) (string, error) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return "", errors.New("invalid remote peer")
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return "", errors.New("invalid remote peer")
	}
	peer = peer.Unmap()
	trustedPeer := false
	for _, prefix := range trusted {
		if prefix.IsValid() && prefix.Contains(peer) {
			trustedPeer = true
			break
		}
	}
	if !trustedPeer {
		return peer.String(), nil
	}
	if forwardedFor == "" || len(forwardedFor) > 512 {
		return "", errors.New("trusted proxy forwarding header unavailable")
	}
	parts := strings.Split(forwardedFor, ",")
	if len(parts) > 16 {
		return "", errors.New("trusted proxy chain too long")
	}
	chain := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		address, err := netip.ParseAddr(value)
		if err != nil || address.Zone() != "" {
			return "", errors.New("invalid forwarded address")
		}
		chain = append(chain, address.Unmap())
	}
	for index := len(chain) - 1; index >= 0; index-- {
		address := chain[index]
		isTrusted := false
		for _, prefix := range trusted {
			if prefix.IsValid() && prefix.Contains(address) {
				isTrusted = true
				break
			}
		}
		if !isTrusted {
			return address.String(), nil
		}
	}
	return "", errors.New("proxy chain contains no untrusted client")
}

type Runtime struct {
	Service     *identity.Service
	Origin      string
	Development bool
}

func NewRuntime(service *identity.Service, origin string, development bool) *Runtime {
	return &Runtime{Service: service, Origin: origin, Development: development}
}
func OfflineRuntime() *Runtime { return &Runtime{} }

// DeliverySender is composed by the Core runtime, never a capture/config toggle.
// Application commands provide a transient delivery only AFTER their transaction
// commits; this adapter cannot open, commit or roll back a business transaction.
type DeliverySender struct {
	renderer  *email.Renderer
	transport email.Transport
}

func NewDeliverySender(origin string, development bool, key, from string, transport http.RoundTripper) (*DeliverySender, error) {
	r, err := email.NewRenderer(origin, development)
	if err != nil {
		return nil, err
	}
	s, err := email.NewResend(key, from, transport)
	if err != nil {
		return nil, err
	}
	return &DeliverySender{r, s}, nil
}

var _ identity.DeliverySender = (*DeliverySender)(nil)

func (s *DeliverySender) Send(ctx context.Context, delivery identity.ActionDelivery) error {
	if s == nil {
		return email.ErrNotSent
	}
	message, err := s.renderer.Render(email.Purpose(delivery.Purpose), email.TemplateData{DisplayName: delivery.DisplayName, ActionURL: delivery.URL, ExpiresAt: delivery.ExpiresAt})
	if err != nil {
		return email.ErrNotSent
	}
	return s.transport.Send(ctx, email.SendInput{AttemptID: delivery.AttemptID, Recipient: delivery.Recipient, Message: message})
}
