package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

type failingLocker struct{}

func (failingLocker) SessionLock(context.Context, *sql.Conn) error { return errors.New("lock failure") }
func (failingLocker) SessionUnlock(context.Context, *sql.Conn) error {
	return errors.New("unlock failure")
}

type lockConnector struct{ closed *bool }

func (c lockConnector) Connect(context.Context) (driver.Conn, error) { return lockConn{c.closed}, nil }
func (lockConnector) Driver() driver.Driver                          { return lockDriver{} }

type lockDriver struct{}

func (lockDriver) Open(string) (driver.Conn, error) { return nil, errors.New("unused") }

type lockConn struct{ closed *bool }

func (lockConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }
func (c lockConn) Close() error                      { *c.closed = true; return nil }
func (lockConn) Begin() (driver.Tx, error)           { return nil, errors.New("unused") }

func TestLockFailureDiscardsUncertainSession(t *testing.T) {
	for _, unlock := range []bool{false, true} {
		closed := false
		pool := sql.OpenDB(lockConnector{&closed})
		defer pool.Close()
		conn, err := pool.Conn(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		locker := boundedLocker{failingLocker{}}
		if unlock {
			err = locker.SessionUnlock(context.Background(), conn)
		} else {
			err = locker.SessionLock(context.Background(), conn)
		}
		if err == nil || !closed {
			t.Fatal("uncertain advisory session returned to reusable pool")
		}
		conn.Close()
	}
}
