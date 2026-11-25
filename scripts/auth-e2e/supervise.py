"""Linux-only bounded Auth check owner. Never prints raw child output.

Subreaping is essential: after the Go launcher/test dies, even setsid children
become OUR children. pidfds target only those owned identities, never stale
numeric process groups. No signalling based on an on-disk PID receipt.
"""
import argparse
import ctypes
import json
import os
from pathlib import Path
import re
import signal
import shutil
import socket
import subprocess
import sys
import time


def write_json(path, value):
    with open(path, "w", encoding="utf-8") as stream:
        json.dump(value, stream)
        stream.flush()
        os.fsync(stream.fileno())


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("recovery", type=Path, help="fresh private /tmp/opencode/auth-* directory")
    parser.add_argument("timeout_ms", type=int, help="positive child deadline in milliseconds")
    parser.add_argument("command", nargs=argparse.REMAINDER, help="owned Auth test command; never a deployed service")
    args = parser.parse_args(argv)
    if args.timeout_ms <= 0 or args.timeout_ms > 1300000 or not args.command:
        parser.error("bounded deadline and owned child command required")
    os.umask(0o077)
    recovery = args.recovery.resolve()
    if not str(recovery).startswith("/tmp/opencode/auth-") or not recovery.is_dir():
        raise RuntimeError("owned recovery directory required")
    manifest = recovery / "ownership.jsonl"
    manifest.touch(mode=0o600, exist_ok=False)
    # Short enough for Chromium's nested Unix-domain singleton socket (<108B).
    resources = recovery / "r"
    resources.mkdir(mode=0o700)
    journal_broken = False

    def record(value):
        nonlocal journal_broken
        try:
            with open(manifest, "a", encoding="utf-8") as stream:
                stream.write(json.dumps(value) + "\n")
                stream.flush()
                os.fsync(stream.fileno())
        except OSError:
            # Do not skip signalling/reaping acquired children on disk failure.
            # Stop work immediately and retain cleanup-unknown evidence.
            journal_broken = True

    # Fail before starting any resource when kernel ownership primitives are absent.
    if ctypes.CDLL(None, use_errno=True).prctl(36, 1, 0, 0, 0) != 0:
        raise RuntimeError("subreaper unavailable")
    fd = os.pidfd_open(os.getpid())
    os.close(fd)
    interrupted = False

    def interrupted_handler(_sig, _frame):
        nonlocal interrupted
        interrupted = True

    signal.signal(signal.SIGTERM, interrupted_handler)
    signal.signal(signal.SIGINT, interrupted_handler)
    env = dict(os.environ)
    env["AUTH_OWNERSHIP_MANIFEST"] = str(manifest)
    deadline = time.monotonic() + args.timeout_ms / 1000
    command = args.command
    after_ready = command[0] == "--timeout-after-ready"
    if after_ready:
        command = command[1:]
    status, timed_out, cleanup = None, False, "unknown"
    issues = []
    with open(recovery / "stdout", "wb") as out, open(recovery / "stderr", "wb") as err:
        child = subprocess.Popen(command, env=env, stdout=out, stderr=err)
        record({"kind": "supervised-child", "pid": child.pid})
        while child.poll() is None and time.monotonic() < deadline and not interrupted and not journal_broken:
            if after_ready and (recovery / "ready").exists():
                deadline = min(deadline, time.monotonic() + 0.2)
                after_ready = False
            time.sleep(0.02)
        status = child.poll()
        timed_out = status is None and not interrupted and not journal_broken
        if status is None:
            # Popen has not reaped this direct child; pidfd prevents PID reuse.
            fd = os.pidfd_open(child.pid)
            try:
                signal.pidfd_send_signal(fd, signal.SIGKILL)
            finally:
                os.close(fd)
            status = child.wait(timeout=5)

        # After launcher exit, only our reparented descendants can be signalled.
        # Killing each adopted parent recursively adopts its descendants. Do not
        # reap until after pidfd signalling; never follow IDs from a receipt.
        cleanup_deadline = time.monotonic() + 45
        end = time.monotonic() + 10
        while time.monotonic() < end:
            children = Path(f"/proc/self/task/{os.getpid()}/children").read_text().split()
            if not children:
                cleanup = "confirmed"
                break
            for raw in children:
                pid = int(raw)
                try:
                    fd = os.pidfd_open(pid)
                except ProcessLookupError:
                    continue
                try:
                    stat = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()
                    if int(stat[1]) != os.getpid():
                        raise RuntimeError("child ownership changed")
                    record({"kind": "adopted-child", "pid": pid, "start": stat[19]})
                    signal.pidfd_send_signal(fd, signal.SIGKILL)
                    os.waitpid(pid, os.WNOHANG)
                finally:
                    os.close(fd)
            time.sleep(0.02)

    # Go fsyncs the TC session intent BEFORE postgres.Run, covering the window
    # before a container handle exists (including Ryuk startup). Record-only
    # container IDs never confer removal authority: inspect exact session labels.
    events = []
    try:
        lines = manifest.read_text().splitlines()
    except OSError:
        lines = []
        journal_broken = True
    for line in lines:
        try:
            events.append(json.loads(line))
        except ValueError:
            journal_broken = True
    if journal_broken:
        cleanup = "unknown"
        issues.append("acquisition journal incomplete")
    sessions = {e["id"] for e in events if e.get("kind") == "postgres-session"}
    # A crash during an in-flight Docker acquisition cannot be claimed fully
    # known just because the current snapshot is empty. Clean discovered owned
    # resources, but retain HOME/evidence and fail cleanup-unknown for that gap.
    if (sum(e.get("kind") == "postgres-session" for e in events) >
            sum(e.get("kind") == "postgres-returned" for e in events)):
        cleanup = "unknown"
        issues.append("in-flight PostgreSQL acquisition")

    def docker(*args):
        remaining = cleanup_deadline - time.monotonic()
        if remaining <= 0:
            raise RuntimeError("cleanup deadline exhausted")
        return subprocess.run(["docker", *args], capture_output=True, text=True, timeout=min(3, remaining),
                              env={"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "HOME": str(recovery)})

    for session in sessions:
        if not re.fullmatch(r"[0-9a-f]{64}", session):
            cleanup = "unknown"
            continue
        end = time.monotonic() + 20
        while True:
            result = docker("ps", "-aq", "--no-trunc", "--filter", f"label=org.testcontainers.sessionId={session}")
            if result.returncode != 0:
                cleanup = "unknown"
                issues.append("container discovery unavailable")
                break
            ids = result.stdout.split()
            for container in ids:
                record({"kind": "discovered-container", "id": container, "session": session})
            if not ids:
                break
            # Give enabled Ryuk the first opportunity. On failure explicitly remove
            # only fully inspected exact-session containers; Ryuk goes LAST.
            if time.monotonic() < end:
                time.sleep(0.2)
                continue
            owned = []
            for container in ids:
                inspected = docker("inspect", "--format", "{{json .Config.Labels}}", container)
                if inspected.returncode != 0:
                    continue  # Ryuk may have just removed it; requery below.
                labels = json.loads(inspected.stdout)
                if not re.fullmatch(r"[0-9a-f]{64}", container) or labels.get("org.testcontainers.sessionId") != session:
                    raise RuntimeError("container ownership mismatch")
                owned.append((labels.get("org.testcontainers.ryuk") == "true", container))
            for _, container in sorted(owned):
                record({"kind": "supervisor-container", "id": container, "session": session})
                docker("rm", "-f", container)
            verified = docker("ps", "-aq", "--filter", f"label=org.testcontainers.sessionId={session}")
            if verified.returncode != 0 or verified.stdout.strip():
                cleanup = "unknown"
                issues.append("session container absence unconfirmed")
            break

    # Rebind every acquisition-time listener, even when success logs are absent.
    for event in events:
        if event.get("kind") == "listener":
            try:
                with socket.socket() as listener:
                    # Match Go net.Listen/Node exclusive listener semantics:
                    # permit TIME_WAIT, never an existing live bound listener.
                    listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                    listener.bind(("127.0.0.1", int(event["id"])))
                    listener.listen()
            except OSError:
                cleanup = "unknown"
                issues.append("listener absence unconfirmed: " + event["id"])
    if time.monotonic() >= cleanup_deadline:
        cleanup = "unknown"
        issues.append("cleanup deadline exhausted")
    if cleanup == "confirmed":
        # This directory was created by us before acquisition, not inferred
        # from a journal path. Failure evidence lives OUTSIDE it and is retained.
        shutil.rmtree(resources)
    write_json(recovery / "result.json", {"status": status, "timedOut": timed_out,
                                          "interrupted": interrupted, "cleanup": cleanup, "issues": issues})


if __name__ == "__main__":
    try:
        main()
    except Exception:
        # No traceback/command/output: arguments may contain fake credentials too.
        sys.stderr.write("Auth supervisor failed; cleanup unknown; retain recovery evidence\n")
        sys.exit(1)
