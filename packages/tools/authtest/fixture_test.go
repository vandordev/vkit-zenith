package authtest

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// This test executable is the only process probe. The inherited socket is
// already bound, so cancellation/crash tests never race a close/rebind handoff.
func TestFixtureOwnedChild(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "auth-owned-child" {
		return
	}
	listener, err := net.FileListener(os.NewFile(3, "owned-loopback"))
	if err != nil {
		os.Exit(3)
	}
	defer listener.Close()
	signal.Ignore(syscall.SIGTERM)
	if err := os.WriteFile("ready", []byte("ready"), 0600); err != nil {
		os.Exit(4)
	}
	if _, err := os.Stat("crash"); err == nil {
		os.Exit(7)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			os.Exit(5)
		}
		conn.Close()
	}
}

func TestAuthFixtureChildOutputCustody(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Fatal("required fixture shell unavailable")
	}
	for _, scope := range []string{"fixture", "recovery"} {
		for _, exit := range []string{"0", "7"} {
			t.Run(scope+"/"+exit, func(t *testing.T) {
				root := t.TempDir()
				if scope == "recovery" {
					root, err = os.MkdirTemp("/tmp/opencode", "auth-output-")
					if err != nil {
						t.Fatal("private recovery root unavailable")
					}
					t.Cleanup(func() {
						if err := os.RemoveAll(root); err != nil {
							t.Error("private recovery root cleanup failed")
						}
						if _, err := os.Stat(root); !os.IsNotExist(err) {
							t.Error("private recovery root remains")
						}
					})
				}
				fixtureRoot := filepath.Join(root, "fixture")
				if err := os.Mkdir(fixtureRoot, 0700); err != nil {
					t.Fatal("owned fixture root unavailable")
				}
				logRoot := fixtureRoot
				journal := ""
				if scope == "recovery" {
					journal = filepath.Join(root, "ownership.jsonl")
					logRoot = root
					if err := os.WriteFile(journal, nil, 0600); err != nil {
						t.Fatal("private ownership journal unavailable")
					}
				}
				t.Setenv("AUTH_OWNERSHIP_MANIFEST", journal)
				completed := filepath.Join(fixtureRoot, "completed")
				command := "printf stdout-private; printf stderr-private >&2; printf done > " + completed + "; exit " + exit
				child, err := StartOwned(context.Background(), fixtureRoot, shell, []string{"-c", command}, nil)
				if err != nil {
					t.Fatal(err)
				}
				deadline := time.Now().Add(3 * time.Second)
				for {
					if _, err := os.Stat(completed); err == nil {
						break
					}
					if time.Now().After(deadline) {
						_ = child.Close()
						t.Fatal("private output child did not complete")
					}
					time.Sleep(10 * time.Millisecond)
				}
				expectedStatus, _ := strconv.Atoi(exit)
				status, waitErr := child.WaitForExit(context.Background())
				if waitErr != nil || status != expectedStatus {
					t.Fatal("owned output child natural status or cleanup failed")
				}
				logs, err := filepath.Glob(filepath.Join(logRoot, "child-*.log"))
				if err != nil || len(logs) != 1 {
					t.Fatal("private child output log missing")
				}
				info, err := os.Stat(logs[0])
				if err != nil || info.Mode().Perm() != 0600 {
					t.Fatal("private child output permissions incorrect")
				}
				body, err := os.Open(logs[0])
				if err != nil {
					t.Fatal("private child output unreadable")
				}
				data, readErr := io.ReadAll(body)
				closeErr := body.Close()
				if readErr != nil || closeErr != nil || !strings.Contains(string(data), "stdout-private") || !strings.Contains(string(data), "stderr-private") {
					t.Fatal("child stdout/stderr not retained privately")
				}
				expectedExit, _ := strconv.Atoi(exit)
				if child.cmd.ProcessState.ExitCode() != expectedExit {
					t.Fatal("child exit status changed during cleanup")
				}
			})
		}
	}
}

func TestAuthFixtureProcessLifecycle(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"startup-cancel", "ignore-term", "assertion-failure", "child-crash", "start-failure"} {
		t.Run(mode, func(t *testing.T) {
			root, err := ownedRoot("auth-process-")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.RemoveAll(root); err != nil {
					t.Error("owned process root cleanup failed")
				}
				if _, err := os.Stat(root); !os.IsNotExist(err) {
					t.Error("owned process root remains")
				}
			}()
			listener, err := ReserveLoopback()
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			file, err := listener.(*net.TCPListener).File()
			if err != nil {
				listener.Close()
				t.Fatal(err)
			}
			defer file.Close()
			defer listener.Close()
			if mode == "child-crash" {
				if err := os.WriteFile(filepath.Join(root, "crash"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			path := executable
			if mode == "start-failure" {
				path = filepath.Join(root, "absent")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			p, err := StartOwned(ctx, root, path, []string{"-test.run=^TestFixtureOwnedChild$", "auth-owned-child"}, []*os.File{file})
			file.Close()
			listener.Close()
			if mode == "start-failure" {
				if err == nil {
					_ = p.Close()
					t.Fatal("start failure accepted")
				}
				check, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
				if err != nil {
					t.Fatal("failed start leaked port")
				}
				check.Close()
				t.Logf("cleanup PASS mode=%s port=%d no-child-acquired", mode, port)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := p.Close(); err != nil {
					t.Error(err)
				}
			}()
			if mode != "startup-cancel" {
				deadline := time.Now().Add(3 * time.Second)
				for {
					_, err := os.Stat(filepath.Join(root, "ready"))
					if err == nil {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("child readiness timed out")
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			if mode != "child-crash" {
				cancel() // cleanup must remain independent from the cancelled work.
			}
			if mode == "assertion-failure" {
				marker := errors.New("injected assertion failure")
				result := func() (err error) {
					defer func() { err = errors.Join(err, p.Close()) }()
					return marker
				}()
				if result == nil || result.Error() != marker.Error() {
					t.Fatal("assertion failure or cleanup lost")
				}
			} else if mode == "child-crash" {
				status, err := p.WaitForExit(ctx)
				if err != nil || status != 7 {
					t.Fatal("owned failed child status or cleanup was lost")
				}
				cancel()
			} else if err := p.Close(); err != nil {
				t.Fatal(err)
			}
			if live, err := liveGroup(p.PID()); err != nil || live {
				t.Fatal("owned child group remains")
			}
			if mode == "child-crash" && p.cmd.ProcessState.ExitCode() != 7 {
				t.Fatal("injected child crash not observed")
			}
			check, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
			if err != nil {
				t.Fatal("owned port not released")
			}
			check.Close()
			t.Logf("cleanup PASS mode=%s group=%d port=%d live-owned=0", mode, p.PID(), port)
		})
	}
}

func TestAuthFixtureEnvironment(t *testing.T) {
	t.Setenv("DB_DSN", "fake-sensitive")
	t.Setenv("PGPASSWORD", "fake-sensitive")
	want := []string{"PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin", "HOME=/owned", "TMPDIR=/owned", "LANG=C.UTF-8"}
	got := ChildEnvironment("/owned")
	if len(got) != len(want) {
		t.Fatal("environment widened")
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatal("environment inherited ambient values")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := StartOwned(ctx, "/owned", "/missing", nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled start accepted")
	}
}
