package staff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

// CanonicalSource accepts an adapter-derived IP only, never a forwarding chain.
// IPv4-mapped IPv6 is IPv4; IPv6 clients share a /64 budget.
func CanonicalSource(source string) (string, error) {
	ip, err := netip.ParseAddr(source)
	if err != nil || ip.Zone() != "" {
		return "", ErrValidation
	}
	ip = ip.Unmap()
	if ip.Is6() {
		return netip.PrefixFrom(ip, 64).Masked().String(), nil
	}
	return ip.String(), nil
}

type attemptBucket struct {
	dimension string
	hash      [32]byte
	limit     int64
}
type admission struct {
	purpose    string
	window, at time.Time
	buckets    []attemptBucket
}

func (s *Service) admit(ctx context.Context, purpose, email, source string) (admission, error) {
	var a admission
	canonical, err := CanonicalSource(source)
	if err != nil {
		return a, err
	}
	a.purpose = purpose
	// Stable deterministic dimension order also governs failure-marker row locks.
	if email != "" {
		a.buckets = append(a.buckets, attemptBucket{"account", sha256.Sum256([]byte(purpose + "/account/" + email)), 10})
	}
	a.buckets = append(a.buckets, attemptBucket{"source", sha256.Sum256([]byte(purpose + "/source/" + canonical)), 100})
	var limited bool
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s';SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		if err := q.StaffLimiterLock(ctx); err != nil {
			return err
		}
		a.at = stamp(s.clock()).Time
		a.window = a.at.Truncate(15 * time.Minute)
		if err := q.StaffLimiterCleanup(ctx, stamp(a.at)); err != nil {
			return err
		}
		count, err := q.StaffLimiterCount(ctx)
		if err != nil {
			return err
		}
		missing := int64(0)
		for _, b := range a.buckets {
			n, err := q.StaffLimiterBucket(ctx, persistence.StaffLimiterBucketParams{Purpose: purpose, Dimension: b.dimension, KeyHash: b.hash[:], WindowStart: stamp(a.window)})
			if errors.Is(err, pgx.ErrNoRows) {
				missing++
			} else if err != nil {
				return err
			}
			if n >= b.limit {
				limited = true
			}
		}
		if count+missing > 20000 {
			limited = true
			return nil
		}
		for _, b := range a.buckets {
			if err := q.StaffLimiterIncrement(ctx, persistence.StaffLimiterIncrementParams{Purpose: purpose, Dimension: b.dimension, KeyHash: b.hash[:], WindowStart: stamp(a.window), ExpiresAt: stamp(a.window.Add(15 * time.Minute))}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return a, safeError(err)
	}
	if limited {
		return a, &RateLimitError{RetryAfter: int(a.window.Add(15*time.Minute).Sub(a.at).Seconds()) + boolSecond(a.at)}
	}
	return a, nil
}

func boolSecond(at time.Time) int {
	if at.Nanosecond() != 0 {
		return 1
	}
	return 0
}

// Failure state and audit are one transaction. No recursive audit/retry on any
// error, no plaintext account/source, no invented staff actor. At most two rows
// per admitted attempt, and once per live bucket/window across all replicas.
func (s *Service) failure(ctx context.Context, a admission) error {
	return safeError(db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s';SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		for _, b := range a.buckets {
			_, err := q.StaffLimiterFailure(ctx, persistence.StaffLimiterFailureParams{Purpose: a.purpose, Dimension: b.dimension, KeyHash: b.hash[:], WindowStart: stamp(a.window)})
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			id, err := newID()
			if err != nil {
				return err
			}
			target := strings.Join([]string{a.purpose, b.dimension, hex.EncodeToString(b.hash[:]), a.window.UTC().Format(time.RFC3339)}, "/")
			if err = q.StaffFailureAudit(ctx, persistence.StaffFailureAuditParams{ID: id, OccurredAt: stamp(a.at), Action: "staff." + a.purpose, TargetID: pgtype.Text{String: target, Valid: true}}); err != nil {
				return err
			}
		}
		return nil
	}))
}
