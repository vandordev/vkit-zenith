package authtest

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// OwnedProcess retains its unreaped leader until cleanup. The leader PID cannot
// be reused before Wait, so only its newly created process group is signalled.
// Children must not escape this group; production daemons are not probe children.
type OwnedProcess struct {
	cmd      *exec.Cmd
	done     bool
	closeErr error
}

func StartOwned(ctx context.Context, root, executable string, args []string, files []*os.File) (*OwnedProcess, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd := exec.Command(executable, args...)
	cmd.Dir, cmd.Env, cmd.ExtraFiles = root, ChildEnvironment(root), files
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Never relay child output: it can contain fake credentials/action URLs too.
	outputRoot := root
	if journal := os.Getenv("AUTH_OWNERSHIP_MANIFEST"); journal != "" {
		outputRoot = filepath.Dir(journal)
	}
	output, err := os.CreateTemp(outputRoot, "child-*.log")
	if err != nil {
		return nil, errors.New("auth fixture: private child output unavailable")
	}
	defer output.Close()
	cmd.Stdout, cmd.Stderr = output, output
	if err := cmd.Start(); err != nil {
		return nil, errors.New("auth fixture: owned child start failed")
	}
	p := &OwnedProcess{cmd: cmd}
	if err := recordOwnership("process", strconv.Itoa(p.PID())); err != nil {
		return nil, errors.Join(err, p.Close())
	}
	return p, nil
}

func (p *OwnedProcess) PID() int { return p.cmd.Process.Pid }

// WaitForExit observes natural leader exit before reaping it, then cleans and
// verifies descendants while the unreaped leader still pins the group ID.
func (p *OwnedProcess) WaitForExit(ctx context.Context) (int, error) {
	if p.done {
		if p.cmd.ProcessState == nil {
			return -1, p.closeErr
		}
		return p.cmd.ProcessState.ExitCode(), p.closeErr
	}
	pid := p.PID()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		exited, err := exitedUnreaped(pid)
		if err != nil {
			return -1, err
		}
		if exited {
			break
		}
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-ticker.C:
		}
	}
	live, err := liveGroup(pid)
	if err != nil {
		return -1, err
	}
	if live {
		if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
			return -1, errors.New("auth fixture: owned group termination failed")
		}
		deadline := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
			return -1, errors.New("auth fixture: owned group forced termination failed")
		}
	}
	// No later call may signal this numeric identity once reaping can begin.
	p.done = true
	_ = p.cmd.Wait()
	status := p.cmd.ProcessState.ExitCode()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		live, err := liveGroup(pid)
		if err != nil {
			p.closeErr = err
			return status, err
		}
		if !live {
			p.closeErr = nil
			return status, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	p.closeErr = errors.New("auth fixture: owned process group remains live")
	return status, p.closeErr
}

func (p *OwnedProcess) Close() (err error) {
	if p.done {
		return p.closeErr
	}
	defer func() {
		if p.done {
			p.closeErr = err
		}
	}()
	pid := p.PID()
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return errors.New("auth fixture: owned group termination failed")
	}
	// Keep the leader unreaped while signalling the group, even if it exits first.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		return errors.New("auth fixture: owned group forced termination failed")
	}
	p.done = true
	waited := make(chan struct{})
	go func() { _ = p.cmd.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(5 * time.Second):
		return errors.New("auth fixture: owned leader could not be reaped")
	}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		live, err := liveGroup(pid)
		if err != nil {
			p.closeErr = err
			return err
		}
		if !live {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("auth fixture: owned process group remains live")
}
