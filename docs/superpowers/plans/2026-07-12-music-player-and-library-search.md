# Music Player and Library Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add in-app music playback (driven by `mpv` over its JSON IPC socket) to Shadowbox, surfaced as a color-sweep progress indicator across the existing ASCII banner, plus type-ahead search in the Library screen, without changing the existing single-screen menu-driven UI structure.

**Architecture:** A new `internal/player` package owns one long-lived `mpv` subprocess per session, controlled over its JSON IPC socket (Unix domain socket on Linux/macOS, named pipe on Windows via `go-winio`). `internal/ui` polls `player.State()` on the existing spinner tick to drive the banner sweep, and forwards a small set of global keys (space/n/p/s/arrows) to the player — but only on screens that aren't capturing free text (Library's new filter, the URL/search Input screen, Settings' inline edit), to avoid the keystroke-stealing bug class that caused real problems in a previous UI redesign attempt.

**Tech Stack:** Go 1.25, existing `charmbracelet/bubbletea`/`bubbles`/`lipgloss`, `mpv` (new external dependency, playback-only, checked lazily), `github.com/Microsoft/go-winio` (new Go dependency, Windows named-pipe support only).

## Global Constraints

- `mpv` is required only for playback, not for the app to run — checked lazily the first time a track is played, never at startup (mirrors how `aria2c` is already optional today via `download.HasAria2()`).
- No new required external dependency beyond `mpv` for playback; no new Go dependency beyond `github.com/Microsoft/go-winio` (Windows-only, behind a build tag).
- Global playback keys (space, `n`, `p`, `s`, left/right arrows) must NOT fire on `screenLibrary`, `screenInput`, or `screenSettingEdit` — those screens capture the same keys for text/filter entry or their own navigation. This was a real, serious bug class in a previous redesign attempt; do not reintroduce it.
- Every task must leave `make lint` and `make test` (or `go vet ./...` + `go test ./...` if `make` targets don't apply to a given task, e.g. the Windows-only file) passing before its commit message is written. Per current project rules, do NOT run `git commit` — write out the exact commit message and stop; the user commits it themselves.
- Tests must be hermetic: no test may depend on a real `mpv` binary being installed, and any test touching `internal/config`/`internal/log` file I/O must sandbox via `t.Setenv("XDG_CONFIG_HOME", t.TempDir())` as its first statement(s) — a real incident already overwrote the developer's actual config file earlier this project from skipping this.
- Match existing code style exactly: `internal/ui` uses value-receiver `model` and lowercase unexported `styles`/`Theme` fields already defined in `themes.go`; `internal/download` uses a flat, un-abstracted `os/exec` style with no interface wrapper around command execution — don't introduce one there.

---

## File Structure

```
internal/player/
├── ipc.go              # NEW: mpv JSON IPC protocol client over any net.Conn
├── ipc_test.go          # NEW
├── socket_unix.go       # NEW: build tag !windows — Unix domain socket path + dial
├── socket_unix_test.go  # NEW
├── socket_windows.go    # NEW: build tag windows — named pipe path + dial (go-winio)
├── player.go            # NEW: Player type — process lifecycle, queue, controls, auto-advance
└── player_test.go       # NEW

internal/download/
└── deps.go               # MODIFIED: register mpv as an optional dependency

internal/ui/
├── bannersweep.go        # NEW: color-sweep banner renderer + now-playing line
├── bannersweep_test.go   # NEW
├── library.go             # MODIFIED: type-ahead filter, Enter-on-track starts playback
├── program.go             # MODIFIED: global playback key routing, player field + cleanup
└── ui_test.go              # MODIFIED: new tests for library filter, playback key routing

README.md                  # MODIFIED: document mpv as a new dependency for playback
docs/INSTALL_FROM_SOURCE.md # MODIFIED: same
```

---

### Task 1: `internal/player` — mpv JSON IPC protocol client

**Files:**
- Create: `internal/player/ipc.go`
- Create: `internal/player/ipc_test.go`

**Interfaces:**
- Produces: `ipcEvent{Event, Reason string}`, `newIPCClient(conn net.Conn) *ipcClient`, `(*ipcClient).command(timeout time.Duration, args ...any) (json.RawMessage, error)`, `(*ipcClient).Events() <-chan ipcEvent`, `(*ipcClient).close() error`.

This is the protocol layer only — it knows nothing about mpv processes or sockets, just how to speak mpv's [JSON IPC protocol](https://mpv.io/manual/stable/#json-ipc) (one JSON object per line; commands get a reply line; mpv may push unsolicited `{"event": "..."}` lines at any time) over any `net.Conn`. Tested with `net.Pipe()` and a fake goroutine standing in for mpv — no real socket or process needed.

- [ ] **Step 1: Write the failing test**

```go
// internal/player/ipc_test.go
package player

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"
)

// fakeMpv reads commands off conn and lets the test decide how to respond,
// simulating the other end of the IPC socket without a real mpv process.
func fakeMpv(t *testing.T, conn net.Conn, respond func(cmd map[string]any) any) {
	t.Helper()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		var cmd map[string]any
		if err := json.Unmarshal(sc.Bytes(), &cmd); err != nil {
			t.Errorf("fakeMpv: bad JSON from client: %v", err)
			return
		}
		reply := respond(cmd)
		data, err := json.Marshal(reply)
		if err != nil {
			t.Errorf("fakeMpv: marshal reply: %v", err)
			return
		}
		if _, err := conn.Write(append(data, '\n')); err != nil {
			return // client closed, nothing more to do
		}
	}
}

func TestCommandReturnsData(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go fakeMpv(t, serverConn, func(cmd map[string]any) any {
		return map[string]any{"data": 42.5, "error": "success", "request_id": cmd["request_id"]}
	})

	c := newIPCClient(clientConn)
	defer c.close()

	data, err := c.command(time.Second, "get_property", "time-pos")
	if err != nil {
		t.Fatalf("command() error = %v", err)
	}
	var got float64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal reply data: %v", err)
	}
	if got != 42.5 {
		t.Errorf("got %v, want 42.5", got)
	}
}

func TestCommandReturnsErrorOnMpvFailure(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go fakeMpv(t, serverConn, func(cmd map[string]any) any {
		return map[string]any{"error": "property not found", "request_id": cmd["request_id"]}
	})

	c := newIPCClient(clientConn)
	defer c.close()

	if _, err := c.command(time.Second, "get_property", "nonexistent"); err == nil {
		t.Fatal("expected an error when mpv reports failure, got nil")
	}
}

func TestCommandTimesOutWithNoReply(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	// serverConn: nobody reads or replies — command() must time out, not hang forever.

	c := newIPCClient(clientConn)
	defer c.close()

	_, err := c.command(50*time.Millisecond, "get_property", "time-pos")
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
}

func TestEventsChannelReceivesUnsolicitedEvents(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newIPCClient(clientConn)
	defer c.close()

	go func() {
		_, _ = serverConn.Write([]byte(`{"event":"end-file","reason":"eof"}` + "\n"))
	}()

	select {
	case ev := <-c.Events():
		if ev.Event != "end-file" || ev.Reason != "eof" {
			t.Errorf("got %+v, want end-file/eof", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the event")
	}
}

func TestEventsChannelClosesWhenConnectionCloses(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	c := newIPCClient(clientConn)
	clientConn.Close() // simulate mpv exiting / the connection dying

	select {
	case _, ok := <-c.Events():
		if ok {
			t.Fatal("expected the events channel to be closed (no more values), got a value instead")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Events() to close after the connection closed")
	}
}

func TestEventDoesNotGetMisreadAsACommandReply(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newIPCClient(clientConn)
	defer c.close()

	// An event arrives first, then the real reply — command() must not
	// treat the event line as its answer.
	go func() {
		_, _ = serverConn.Write([]byte(`{"event":"idle"}` + "\n"))
		time.Sleep(20 * time.Millisecond)
		sc := bufio.NewScanner(serverConn)
		sc.Scan()
		var cmd map[string]any
		_ = json.Unmarshal(sc.Bytes(), &cmd)
		reply, _ := json.Marshal(map[string]any{"data": 7.0, "error": "success", "request_id": cmd["request_id"]})
		_, _ = serverConn.Write(append(reply, '\n'))
	}()

	data, err := c.command(time.Second, "get_property", "duration")
	if err != nil {
		t.Fatalf("command() error = %v", err)
	}
	var got float64
	_ = json.Unmarshal(data, &got)
	if got != 7.0 {
		t.Errorf("got %v, want 7.0 (the event must not have been consumed as the reply)", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/player/... -v`
Expected: FAIL — `package player` doesn't exist yet / `newIPCClient` undefined. (6 tests total once `TestEventsChannelClosesWhenConnectionCloses` above is included.)

- [ ] **Step 3: Write the implementation**

```go
// internal/player/ipc.go

// Package player drives audio playback through a long-lived mpv subprocess,
// controlled over its JSON IPC socket. See https://mpv.io/manual/stable/#json-ipc.
package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ipcEvent is an unsolicited message mpv may push at any time, e.g. when a
// file finishes playing.
type ipcEvent struct {
	Event  string `json:"event"`
	Reason string `json:"reason"`
}

// ipcReply is mpv's response to a command we sent.
type ipcReply struct {
	Data      json.RawMessage `json:"data"`
	Error     string          `json:"error"`
	RequestID int64           `json:"request_id"`
}

// ipcClient speaks mpv's JSON IPC protocol over any net.Conn: one JSON
// object per line, a reply line per command, and unsolicited event lines
// mpv can push at any time (even between a command and its reply).
type ipcClient struct {
	conn     net.Conn
	mu       sync.Mutex // serializes command() calls: one outstanding request at a time
	replies  chan ipcReply
	eventsCh chan ipcEvent
	nextID   int64
	closed   chan struct{}
}

func newIPCClient(conn net.Conn) *ipcClient {
	c := &ipcClient{
		conn:     conn,
		replies:  make(chan ipcReply, 1),
		eventsCh: make(chan ipcEvent, 16),
		closed:   make(chan struct{}),
	}
	go c.readLoop()
	return c
}

// readLoop demuxes every line mpv sends into either the events channel
// (unsolicited, has an "event" key) or the replies channel (an answer to a
// command we sent). It runs for the lifetime of the connection; when the
// connection closes (mpv exits, crashes, or we call close()), it closes
// both closed and eventsCh, so Player's watchEvents range loop (Task 4)
// terminates and can tell the difference between "waiting" and "gone."
func (c *ipcClient) readLoop() {
	defer close(c.closed)
	defer close(c.eventsCh)
	sc := bufio.NewScanner(c.conn)
	for sc.Scan() {
		line := sc.Bytes()

		var probe struct {
			Event *string `json:"event"`
		}
		if err := json.Unmarshal(line, &probe); err == nil && probe.Event != nil {
			var ev ipcEvent
			_ = json.Unmarshal(line, &ev)
			select {
			case c.eventsCh <- ev:
			default: // drop if nobody is listening fast enough
			}
			continue
		}

		var reply ipcReply
		if err := json.Unmarshal(line, &reply); err == nil {
			select {
			case c.replies <- reply:
			default:
			}
		}
	}
}

// command sends an mpv command and waits for its reply, up to timeout.
// Only one command may be outstanding at a time (enforced by mu) — this is
// fine for our usage, which never issues concurrent commands.
func (c *ipcClient) command(timeout time.Duration, args ...any) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := atomic.AddInt64(&c.nextID, 1)
	payload, err := json.Marshal(map[string]any{"command": args, "request_id": id})
	if err != nil {
		return nil, err
	}
	// The write itself can block indefinitely if the other end never reads
	// (a hung mpv, or nothing listening at all) — bound it by the same
	// timeout so command() can never hang before even reaching the reply
	// wait below. Discovered during Task 1's own execution: net.Pipe() is
	// fully synchronous, so TestCommandTimesOutWithNoReply hung on this
	// exact gap before this fix existed.
	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("mpv: set write deadline: %w", err)
	}
	if _, err := c.conn.Write(append(payload, '\n')); err != nil {
		return nil, fmt.Errorf("mpv: write command: %w", err)
	}
	if err := c.conn.SetWriteDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("mpv: clear write deadline: %w", err)
	}

	select {
	case reply := <-c.replies:
		if reply.Error != "success" {
			return nil, fmt.Errorf("mpv: %s", reply.Error)
		}
		return reply.Data, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("mpv: command timed out after %s", timeout)
	case <-c.closed:
		return nil, fmt.Errorf("mpv: connection closed")
	}
}

// Events returns the channel of unsolicited events mpv pushes (e.g. a track
// finishing). Never closed while the client is open; safe to range/select on.
func (c *ipcClient) Events() <-chan ipcEvent {
	return c.eventsCh
}

// close shuts down the underlying connection.
func (c *ipcClient) close() error {
	return c.conn.Close()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/player/... -v`
Expected: PASS (6 tests)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/player/...
make lint
```
Expected: both clean. Per project rules, do NOT run `git commit`. Write out this commit message and stop:

```
feat(player): add mpv JSON IPC protocol client

Speaks mpv's line-delimited JSON IPC protocol over any net.Conn:
sends commands and correlates replies, and demuxes unsolicited
events (e.g. end-of-file) onto a separate channel so they can never
be misread as a command's reply.
```

---

### Task 2: `internal/player` — Unix socket path + dial

**Files:**
- Create: `internal/player/socket_unix.go` (build tag `!windows`)
- Create: `internal/player/socket_unix_test.go` (build tag `!windows`)

**Interfaces:**
- Consumes: nothing from Task 1 directly (this is dial plumbing `player.go` will combine with `newIPCClient` in Task 4).
- Produces: `socketPath() string`, `dialSocket(path string, timeout time.Duration) (net.Conn, error)`.

- [ ] **Step 1: Write the failing test**

```go
// internal/player/socket_unix_test.go
//go:build !windows

package player

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDialSocketConnectsOnceListenerIsReady(t *testing.T) {
	path := socketPath()
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer ln.Close()
	defer os.Remove(path)

	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := ln.Accept()
		if err == nil {
			accepted <- c
		}
	}()

	conn, err := dialSocket(path, time.Second)
	if err != nil {
		t.Fatalf("dialSocket() error = %v", err)
	}
	defer conn.Close()

	select {
	case c := <-accepted:
		defer c.Close()
	case <-time.After(time.Second):
		t.Fatal("listener never accepted the connection")
	}
}

func TestDialSocketTimesOutWhenNothingListens(t *testing.T) {
	_, err := dialSocket(socketPath(), 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error dialing a socket nobody is listening on")
	}
}

func TestSocketPathIsUnderTempDirAndUnique(t *testing.T) {
	a, b := socketPath(), socketPath()
	if !strings.Contains(a, os.TempDir()) {
		t.Errorf("socketPath() = %q, want it under %q", a, os.TempDir())
	}
	if a == b {
		t.Errorf("two calls to socketPath() returned the same path %q, want unique paths", a)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/player/... -run TestDialSocket -v` and `-run TestSocketPath -v`
Expected: FAIL — `socketPath`/`dialSocket` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/player/socket_unix.go
//go:build !windows

package player

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// socketPath returns a fresh, unique path for mpv's IPC socket under the OS
// temp directory. Called once per Player, so PID + nanosecond timestamp is
// enough to avoid collisions with any previous run.
func socketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("shadowbox-mpv-%d-%d.sock", os.Getpid(), time.Now().UnixNano()))
}

// dialSocket connects to the mpv IPC socket at path, retrying briefly since
// mpv creates the socket file shortly after starting, not instantly.
func dialSocket(path string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	return nil, fmt.Errorf("dial mpv socket %s: %w", path, lastErr)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/player/... -v`
Expected: PASS (all tests from Task 1 and Task 2, 9 total)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/player/...
make lint
```
Expected: both clean. Write out this commit message and stop:

```
feat(player): add Unix domain socket dial for mpv IPC

Generates a unique per-run socket path under the OS temp dir and
dials it with a short retry loop, since mpv creates the socket file
shortly after starting rather than instantly.
```

---

### Task 3: `internal/player` — Windows named pipe path + dial

**Files:**
- Create: `internal/player/socket_windows.go` (build tag `windows`)
- Modify: `go.mod`, `go.sum` (adds `github.com/Microsoft/go-winio`)

**Interfaces:**
- Produces: `socketPath() string`, `dialSocket(path string, timeout time.Duration) (net.Conn, error)` — same signatures as Task 2, the Windows-only implementation. Exactly one of Task 2's or this task's file is compiled for a given `GOOS`, so `player.go` (Task 4) calls these names without caring which platform it's on.

This file cannot be exercised by `go test` on this (non-Windows) machine — its verification step is a cross-compile check instead of a test run. Note this honestly; it is a real, accepted gap in this project's test coverage for the Windows platform, not an oversight.

- [ ] **Step 1: Add the dependency**

```bash
go get github.com/Microsoft/go-winio
```

- [ ] **Step 2: Write the implementation**

```go
// internal/player/socket_windows.go
//go:build windows

package player

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// socketPath returns a fresh, unique named-pipe path for mpv's IPC server.
// Nanosecond timestamp is enough to avoid collisions with any previous run.
func socketPath() string {
	return fmt.Sprintf(`\\.\pipe\shadowbox-mpv-%d`, time.Now().UnixNano())
}

// dialSocket connects to the mpv IPC named pipe at path, retrying briefly
// since mpv creates the pipe shortly after starting, not instantly.
func dialSocket(path string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := winio.DialPipe(path, nil)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	return nil, fmt.Errorf("dial mpv pipe %s: %w", path, lastErr)
}
```

- [ ] **Step 3: Verify it cross-compiles**

Run: `GOOS=windows GOARCH=amd64 go build ./internal/player/...`
Expected: no output (success). This is the only verification possible for this file on a non-Windows development machine — flag this limitation in the report rather than silently treating a successful cross-compile as equivalent to a real test run.

- [ ] **Step 4: Verify the rest of the module still builds and lints**

```bash
go build ./...
go vet ./...
make lint
```
Expected: all clean.

- [ ] **Step 5: Verify and stop (do not commit)**

Write out this commit message and stop:

```
feat(player): add Windows named-pipe dial for mpv IPC

Windows counterpart to the Unix domain socket dial, using go-winio
for named-pipe support. Only cross-compile-verified on this
machine (GOOS=windows go build) — not test-run, since that requires
an actual Windows environment.
```

---

### Task 4: `internal/player` — Player (process lifecycle, queue, controls, auto-advance)

**Files:**
- Create: `internal/player/player.go`
- Create: `internal/player/player_test.go`

**Interfaces:**
- Consumes: `newIPCClient`, `(*ipcClient).command`, `(*ipcClient).Events`, `(*ipcClient).close` (Task 1); `socketPath`, `dialSocket` (Task 2/3, whichever matches `GOOS`).
- Produces: `player.Track{Path, Title string}`, `player.State{Title string, Elapsed, Duration time.Duration, Playing bool, LastError string}`, `player.Available() bool`, `player.New() (*Player, error)`, `(*Player).Load(tracks []Track, startIndex int) error`, `(*Player).State() State`, `(*Player).TogglePause() error`, `(*Player).Stop() error`, `(*Player).SeekBy(delta time.Duration) error`, `(*Player).Next() error`, `(*Player).Prev() error`, `(*Player).Close() error`.

`Player` depends on the IPC layer through a small unexported interface (`mpvConn`), satisfied by `*ipcClient` in production and a fake in tests — this is what makes the queue/state/auto-advance logic testable without a real `mpv` process, matching this project's existing convention of not unit-testing actual external-process invocation (see `internal/download`), while still meaningfully testing everything above that layer.

- [ ] **Step 1: Write the failing test**

```go
// internal/player/player_test.go
package player

import (
	"encoding/json"
	"testing"
	"time"
)

// fakeConn is a test double for mpvConn — no real socket or process.
type fakeConn struct {
	commands []string // records "arg0 arg1 ..." for each command sent, for assertions
	timePos  float64
	duration float64
	events   chan ipcEvent
}

func newFakeConn() *fakeConn {
	return &fakeConn{events: make(chan ipcEvent, 4)}
}

func (f *fakeConn) command(_ time.Duration, args ...any) (json.RawMessage, error) {
	rec := ""
	for i, a := range args {
		if i > 0 {
			rec += " "
		}
		rec += toStr(a)
	}
	f.commands = append(f.commands, rec)

	if len(args) >= 2 && args[0] == "get_property" {
		switch args[1] {
		case "time-pos":
			return json.Marshal(f.timePos)
		case "duration":
			return json.Marshal(f.duration)
		}
	}
	return json.Marshal(nil)
}

func toStr(a any) string {
	switch v := a.(type) {
	case string:
		return v
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func (f *fakeConn) Events() <-chan ipcEvent { return f.events }

// close simulates the real ipcClient.close(), which (via readLoop exiting)
// closes the events channel — Player.watchEvents relies on this to detect
// both an intentional Close() and an unexpected disconnect the same way.
func (f *fakeConn) close() error {
	close(f.events)
	return nil
}

func testTracks() []Track {
	return []Track{
		{Path: "/music/A/1.opus", Title: "First"},
		{Path: "/music/A/2.opus", Title: "Second"},
		{Path: "/music/A/3.opus", Title: "Third"},
	}
}

func TestLoadSendsLoadfileAndSetsState(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)

	if err := p.Load(testTracks(), 1); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	st := p.State()
	if st.Title != "Second" || !st.Playing {
		t.Errorf("State() = %+v, want Title=Second Playing=true", st)
	}
	if len(conn.commands) == 0 || conn.commands[0] != `loadfile /music/A/2.opus replace` {
		t.Errorf("commands = %v, want first command to load /music/A/2.opus", conn.commands)
	}
}

func TestStatePollsElapsedAndDuration(t *testing.T) {
	conn := newFakeConn()
	conn.timePos = 12.5
	conn.duration = 200
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	st := p.State()
	if st.Elapsed != 12500*time.Millisecond {
		t.Errorf("Elapsed = %v, want 12.5s", st.Elapsed)
	}
	if st.Duration != 200*time.Second {
		t.Errorf("Duration = %v, want 200s", st.Duration)
	}
}

func TestStateIdleWhenNothingLoaded(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)

	st := p.State()
	if st.Playing || st.Title != "" {
		t.Errorf("State() = %+v, want idle zero value before any Load()", st)
	}
	if len(conn.commands) != 0 {
		t.Errorf("expected no IPC calls before any track is loaded, got %v", conn.commands)
	}
}

func TestTogglePauseFlipsPlayingAndSendsSetProperty(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	if err := p.TogglePause(); err != nil {
		t.Fatalf("TogglePause() error = %v", err)
	}
	// Check the command sent BEFORE calling State() below — State() itself
	// issues get_property commands, which would otherwise become the new
	// "last command" and mask what TogglePause() actually sent. (Caught by
	// running this test during Task 4's own execution — the original
	// ordering here made the assertion fail against a real, correctly
	// implemented Player.)
	last := conn.commands[len(conn.commands)-1]
	if last != "set_property pause true" {
		t.Errorf("last command = %q, want pause set to true", last)
	}
	if p.State().Playing {
		t.Error("expected Playing=false after first TogglePause()")
	}

	if err := p.TogglePause(); err != nil {
		t.Fatalf("TogglePause() error = %v", err)
	}
	if !p.State().Playing {
		t.Error("expected Playing=true after second TogglePause()")
	}
}

func TestNextAndPrevMoveWithinQueueBounds(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	if err := p.Next(); err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if p.State().Title != "Second" {
		t.Errorf("after Next(), Title = %q, want Second", p.State().Title)
	}

	if err := p.Prev(); err != nil {
		t.Fatalf("Prev() error = %v", err)
	}
	if p.State().Title != "First" {
		t.Errorf("after Prev(), Title = %q, want First", p.State().Title)
	}

	// Prev() at the start of the queue is a no-op, not an error or wraparound.
	if err := p.Prev(); err != nil {
		t.Fatalf("Prev() at queue start error = %v", err)
	}
	if p.State().Title != "First" {
		t.Errorf("Prev() at queue start changed track to %q, want it to stay First", p.State().Title)
	}

	// Next() past the end of the queue is a no-op.
	_ = p.Next()
	_ = p.Next()
	if err := p.Next(); err != nil {
		t.Fatalf("Next() at queue end error = %v", err)
	}
	if p.State().Title != "Third" {
		t.Errorf("Next() past queue end changed track to %q, want it to stay Third", p.State().Title)
	}
}

func TestStopClearsState(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	if err := p.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	st := p.State()
	if st.Playing || st.Title != "" {
		t.Errorf("State() after Stop() = %+v, want idle zero value", st)
	}
}

func TestSeekByClampsToTrackBounds(t *testing.T) {
	conn := newFakeConn()
	conn.duration = 100
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	if err := p.SeekBy(-30 * time.Second); err != nil {
		t.Fatalf("SeekBy() error = %v", err)
	}
	if p.State().Elapsed < 0 {
		t.Errorf("Elapsed = %v after seeking back from 0, want clamped to >= 0", p.State().Elapsed)
	}
}

func TestEndFileEventAdvancesToNextTrack(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	conn.events <- ipcEvent{Event: "end-file", Reason: "eof"}

	// Auto-advance runs on a background goroutine; poll briefly for it.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if p.State().Title == "Second" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected auto-advance to Second after end-file/eof, got %q", p.State().Title)
}

func TestEndFileEventAtEndOfQueueClearsState(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 2) // last track

	conn.events <- ipcEvent{Event: "end-file", Reason: "eof"}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if p.State().Title == "" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected state to clear after the last track ends, got %q", p.State().Title)
}

func TestEndFileEventWithStopReasonDoesNotAdvance(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	conn.events <- ipcEvent{Event: "end-file", Reason: "stop"} // user-initiated stop, not natural end or an error

	time.Sleep(50 * time.Millisecond) // give the goroutine a chance to (wrongly) act
	if p.State().Title != "First" {
		t.Errorf("a stop-reason end-file must not trigger auto-advance, got %q", p.State().Title)
	}
}

func TestEndFileErrorReasonAdvancesAndSetsLastError(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0) // corrupt/unsupported first track

	conn.events <- ipcEvent{Event: "end-file", Reason: "error"}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		st := p.State()
		if st.Title == "Second" && st.LastError != "" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected auto-advance to Second with LastError set after end-file/error, got %+v", p.State())
}

func TestLoadClearsLastErrorFromAPreviousTrack(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)
	conn.events <- ipcEvent{Event: "end-file", Reason: "error"}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && p.State().LastError == "" {
		time.Sleep(5 * time.Millisecond)
	}
	if p.State().LastError == "" {
		t.Fatal("setup failed: LastError never got set")
	}

	_ = p.Load(testTracks(), 1) // a fresh, successful Load
	if p.State().LastError != "" {
		t.Errorf("LastError = %q after a fresh Load, want cleared", p.State().LastError)
	}
}

func TestConnectionClosingUnexpectedlyClearsStateWithLastError(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	close(conn.events) // simulate the IPC connection dying (mpv crashed)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		st := p.State()
		if st.Title == "" && st.LastError != "" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected state to clear with LastError set after the connection closed, got %+v", p.State())
}

func TestCloseDoesNotSetLastErrorOnIntentionalShutdown(t *testing.T) {
	conn := newFakeConn()
	p := newPlayerWithConn(conn)
	_ = p.Load(testTracks(), 0)

	_ = p.Close() // closes conn.events itself, since fakeConn.close() does so — see fakeConn below

	time.Sleep(50 * time.Millisecond) // give watchEvents a chance to (wrongly) set LastError
	if p.State().LastError != "" {
		t.Errorf("LastError = %q after an intentional Close(), want empty", p.State().LastError)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/player/... -run 'TestLoad|TestState|TestToggle|TestNext|TestPrev|TestStop|TestSeek|TestEndFile' -v`
Expected: FAIL — `Track`, `newPlayerWithConn`, `Player` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/player/player.go
package player

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const ipcTimeout = 2 * time.Second

// Track is one playable file plus the title to show for it — the caller
// (Library) derives Title from the filename so Player never needs to ask
// mpv for metadata.
type Track struct {
	Path  string
	Title string
}

// State is a snapshot of what's currently playing. LastError is set when
// the previous track ended abnormally (a corrupt/unsupported file mpv
// couldn't play, or the mpv connection itself dying) and cleared by the
// next successful Load.
type State struct {
	Title     string
	Elapsed   time.Duration
	Duration  time.Duration
	Playing   bool
	LastError string
}

// mpvConn is the subset of *ipcClient that Player depends on, satisfied by
// *ipcClient in production and a fake in tests.
type mpvConn interface {
	command(timeout time.Duration, args ...any) (json.RawMessage, error)
	Events() <-chan ipcEvent
	close() error
}

// Player owns one mpv subprocess for the session and the current queue.
type Player struct {
	conn mpvConn
	cmd  *exec.Cmd // nil when built via newPlayerWithConn for tests

	mu      sync.Mutex
	queue   []Track
	index   int
	state   State
	closing bool // set by Close() before it tears the connection down, so
	             // watchEvents can tell an intentional shutdown apart from
	             // mpv dying unexpectedly
}

// Available reports whether mpv is installed and on PATH. Callers should
// check this before New() and show a clear message if it's false — mpv is
// required only for playback, never for the rest of the app.
func Available() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

// New spawns an idle mpv process with its IPC socket enabled and connects
// to it. The process plays nothing until the first Load().
func New() (*Player, error) {
	path := socketPath()
	cmd := exec.Command("mpv", "--idle", "--no-video", "--no-terminal", "--input-ipc-server="+path)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start mpv: %w", err)
	}

	conn, err := dialSocket(path, 2*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("connect to mpv: %w", err)
	}

	p := newPlayerWithConn(newIPCClient(conn))
	p.cmd = cmd
	return p, nil
}

// newPlayerWithConn builds a Player around any mpvConn, real or fake, and
// starts the background goroutine that watches for track-end events.
func newPlayerWithConn(conn mpvConn) *Player {
	p := &Player{conn: conn}
	go p.watchEvents()
	return p
}

// watchEvents auto-advances the queue when mpv reports a track ended
// naturally (reason "eof") or couldn't be played (reason "error" — a
// corrupt/unsupported file), setting LastError in the latter case. A
// user-initiated stop (reason "stop") does not advance. When the events
// channel itself closes — mpv exited or crashed — state clears with
// LastError set, unless Close() caused it intentionally.
func (p *Player) watchEvents() {
	for ev := range p.conn.Events() {
		if ev.Event != "end-file" {
			continue
		}
		p.mu.Lock()
		switch ev.Reason {
		case "eof":
			p.advanceOnEndLocked()
		case "error":
			p.advanceOnEndLocked()
			p.state.LastError = "playback error, skipping track"
		}
		p.mu.Unlock()
	}

	p.mu.Lock()
	if !p.closing {
		p.state = State{LastError: "mpv exited unexpectedly"}
	}
	p.mu.Unlock()
}

// Load starts playing tracks[startIndex] and sets the queue used for
// Next()/Prev() and auto-advance.
func (p *Player) Load(tracks []Track, startIndex int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = tracks
	p.index = startIndex
	return p.loadCurrentLocked()
}

func (p *Player) loadCurrentLocked() error {
	if p.index < 0 || p.index >= len(p.queue) {
		p.state = State{}
		return nil
	}
	track := p.queue[p.index]
	if _, err := p.conn.command(ipcTimeout, "loadfile", track.Path, "replace"); err != nil {
		return err
	}
	p.state = State{Title: track.Title, Playing: true}
	return nil
}

// State returns the current playback snapshot, polling mpv for the live
// position and duration when a track is loaded.
func (p *Player) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state.Title == "" {
		return p.state
	}
	if data, err := p.conn.command(ipcTimeout, "get_property", "time-pos"); err == nil {
		var secs float64
		if json.Unmarshal(data, &secs) == nil {
			p.state.Elapsed = time.Duration(secs * float64(time.Second))
		}
	}
	if data, err := p.conn.command(ipcTimeout, "get_property", "duration"); err == nil {
		var secs float64
		if json.Unmarshal(data, &secs) == nil && secs > 0 {
			p.state.Duration = time.Duration(secs * float64(time.Second))
		}
	}
	return p.state
}

// TogglePause flips play/pause.
func (p *Player) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	next := !p.state.Playing
	if _, err := p.conn.command(ipcTimeout, "set_property", "pause", !next); err != nil {
		return err
	}
	p.state.Playing = next
	return nil
}

// Stop halts playback and clears the queue and now-playing state.
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.conn.command(ipcTimeout, "stop"); err != nil {
		return err
	}
	p.queue = nil
	p.index = 0
	p.state = State{}
	return nil
}

// SeekBy seeks relative to the current position; mpv clamps to the track's
// own bounds, and the locally-tracked Elapsed is nudged to match so the
// banner sweep doesn't visibly lag until the next poll.
func (p *Player) SeekBy(delta time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.conn.command(ipcTimeout, "seek", delta.Seconds(), "relative"); err != nil {
		return err
	}
	p.state.Elapsed += delta
	if p.state.Elapsed < 0 {
		p.state.Elapsed = 0
	}
	if p.state.Duration > 0 && p.state.Elapsed > p.state.Duration {
		p.state.Elapsed = p.state.Duration
	}
	return nil
}

// Next skips to the next track in the queue; a no-op past the end.
func (p *Player) Next() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.skipLocked(1)
}

// Prev skips to the previous track in the queue; a no-op before the start.
func (p *Player) Prev() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.skipLocked(-1)
}

// skipLocked moves by delta within the queue for an explicit Next()/Prev()
// call. Out of bounds is a no-op that leaves the current track playing —
// distinct from advanceOnEndLocked, which clears state at the same
// boundary because it means the queue actually finished playing through.
// (Originally one shared advanceLocked function; split during Task 4's own
// execution after TestNextAndPrevMoveWithinQueueBounds caught the two
// callers needing different behavior at the same boundary.)
func (p *Player) skipLocked(delta int) error {
	next := p.index + delta
	if next < 0 || next >= len(p.queue) {
		return nil
	}
	p.index = next
	return p.loadCurrentLocked()
}

// advanceOnEndLocked moves to the next track after mpv reports the current
// one ended (naturally or via error). Past the end of the queue, this
// clears state — the queue is exhausted, unlike an explicit Next() at the
// same position, which is just a no-op.
func (p *Player) advanceOnEndLocked() {
	next := p.index + 1
	if next >= len(p.queue) {
		p.state = State{}
		return
	}
	p.index = next
	_ = p.loadCurrentLocked()
}

// Close stops mpv and releases the IPC connection. Safe to call once when
// the app exits.
func (p *Player) Close() error {
	p.mu.Lock()
	p.closing = true
	p.mu.Unlock()

	_, _ = p.conn.command(ipcTimeout, "quit")
	err := p.conn.close()
	if p.cmd != nil {
		_ = p.cmd.Wait()
	}
	return err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/player/... -v`
Expected: PASS (all tests from Tasks 1, 2, and 4 — 27 total)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/player/...
make lint
```
Expected: both clean. Write out this commit message and stop:

```
feat(player): add Player with queue, controls, and auto-advance

Owns one mpv process for the session; Load/Next/Prev manage a
simple in-memory queue, State() polls mpv for live position and
duration, and a background goroutine auto-advances the queue when
mpv reports a track ended naturally (not on a user-initiated stop),
surfacing LastError and skipping ahead on a corrupt/unsupported
file, and clearing state with LastError set if mpv's connection
drops unexpectedly (vs. an intentional Close()). Depends on the IPC
layer through a small interface so the queue and state logic is
fully testable without a real mpv process.
```

---

### Task 5: `internal/download/deps.go` — register mpv as an optional dependency

**Files:**
- Modify: `internal/download/deps.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `download.Dependencies()` now includes an `mpv` entry (`Required: false`), so `shadowbox doctor` lists it automatically — no changes needed to `internal/cmd/doctor.go`, which already ranges over `Dependencies()` generically.

This does not replace `player.Available()` (Task 4) — that's the lazy, in-the-moment check used right before playback. This entry is purely for `shadowbox doctor`'s informational listing, matching how `aria2c` already has both a `Dependencies()` entry and its own dedicated `download.HasAria2()` check used at the actual point of use.

- [ ] **Step 1: Write the failing test**

```go
// internal/download/deps_test.go
package download

import "testing"

func TestDependenciesIncludesMpvAsOptional(t *testing.T) {
	for _, d := range Dependencies() {
		if d.Name == "mpv" {
			if d.Required {
				t.Error(`mpv dependency has Required = true, want false (playback-only, not needed to run Shadowbox)`)
			}
			return
		}
	}
	t.Fatal("Dependencies() does not include mpv")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/download/... -run TestDependenciesIncludesMpv -v`
Expected: FAIL — no `mpv` entry in `Dependencies()`.

- [ ] **Step 3: Write the implementation**

```go
// internal/download/deps.go — modify the specs slice inside Dependencies()
	specs := []Dependency{
		{Name: "yt-dlp", Purpose: "downloading audio from YouTube/Bandcamp/KHInsider", Required: true},
		{Name: "ffmpeg", Purpose: "audio extraction and conversion", Required: true},
		{Name: "aria2c", Purpose: "accelerated multi-connection downloads", Required: false},
		{Name: "mpv", Purpose: "in-app audio playback", Required: false},
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/download/... -v`
Expected: PASS (all existing `internal/download` tests plus the new one)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/download/...
make lint
make test
```
Expected: all clean. Write out this commit message and stop:

```
feat(download): list mpv as an optional dependency

Surfaces mpv in `shadowbox doctor`'s dependency listing, matching
the existing aria2c pattern — informational only, not a startup
gate; playback itself checks player.Available() lazily.
```

---

### Task 6: `internal/ui` — banner color-sweep renderer + now-playing line

**Files:**
- Create: `internal/ui/bannersweep.go`
- Create: `internal/ui/bannersweep_test.go`

**Interfaces:**
- Consumes: `player.State` (Task 4), the existing `banner` const and `styles`/`Theme` types (`internal/ui/themes.go`).
- Produces: `renderBannerWithPlayback(st styles, theme Theme, state player.State) string` — every screen's `view*()` function will switch to calling this instead of `m.st.title.Render(banner)` directly (wired in Task 8, not here — this task only adds the renderer and its tests).

The banner is fixed ASCII art, one rune wide per column regardless of the multi-byte box-drawing characters used — sweep progress is computed per character position in the string, not per visual column, which is fine here since the banner has no wide (double-width) runes.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/bannersweep_test.go
package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/EnJulian/shadowbox/internal/player"
)

func TestRenderBannerWithPlaybackIdleMatchesPlainBanner(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")

	got := renderBannerWithPlayback(st, theme, player.State{})
	want := st.title.Render(banner)
	if got != want {
		t.Errorf("idle banner should render exactly like the plain banner, got a different string")
	}
}

func TestRenderBannerWithPlaybackShowsNowPlayingLine(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")
	state := player.State{Title: "Feather", Elapsed: 90 * time.Second, Duration: 222 * time.Second, Playing: true}

	got := renderBannerWithPlayback(st, theme, state)
	if !strings.Contains(got, "Feather") {
		t.Error("expected the now-playing line to contain the track title")
	}
	if !strings.Contains(got, "1:30") || !strings.Contains(got, "3:42") {
		t.Errorf("expected elapsed/total time as m:ss, got:\n%s", got)
	}
}

func TestRenderBannerWithPlaybackSweepsProportionally(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")

	early := renderBannerWithPlayback(st, theme, player.State{Title: "X", Elapsed: 1 * time.Second, Duration: 100 * time.Second, Playing: true})
	late := renderBannerWithPlayback(st, theme, player.State{Title: "X", Elapsed: 90 * time.Second, Duration: 100 * time.Second, Playing: true})

	if early == late {
		t.Error("banner rendering should differ between 1% and 90% progress")
	}
}

func TestRenderBannerWithPlaybackHandlesZeroDuration(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")
	// Duration is briefly 0 right after Load(), before the first poll — must
	// not panic or divide by zero.
	got := renderBannerWithPlayback(st, theme, player.State{Title: "X", Playing: true})
	if !strings.Contains(got, "X") {
		t.Error("expected the title to still render with a zero duration")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -run TestRenderBannerWithPlayback -v`
Expected: FAIL — `renderBannerWithPlayback` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/bannersweep.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/EnJulian/shadowbox/internal/player"
)

// renderBannerWithPlayback renders the banner as-is when nothing is
// playing, or with a left-to-right color sweep (muted -> accent) tracking
// playback progress, plus a now-playing title/time line below it.
func renderBannerWithPlayback(st styles, theme Theme, state player.State) string {
	if state.Title == "" {
		return st.title.Render(banner)
	}

	lines := strings.Split(banner, "\n")
	width := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > width {
			width = n
		}
	}

	var fraction float64
	if state.Duration > 0 {
		fraction = float64(state.Elapsed) / float64(state.Duration)
		if fraction > 1 {
			fraction = 1
		}
		if fraction < 0 {
			fraction = 0
		}
	}
	fillTo := int(fraction * float64(width))

	mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	accentStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)

	var b strings.Builder
	for _, line := range lines {
		runes := []rune(line)
		for i, r := range runes {
			if i < fillTo {
				b.WriteString(accentStyle.Render(string(r)))
			} else {
				b.WriteString(mutedStyle.Render(string(r)))
			}
		}
		b.WriteString("\n")
	}

	nowPlaying := fmt.Sprintf("  %s  %s / %s", state.Title, formatDuration(state.Elapsed), formatDuration(state.Duration))
	b.WriteString(st.subtitle.Render(nowPlaying))
	return b.String()
}

// formatDuration renders a duration as m:ss, matching the mockup the design
// was approved against.
func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	if total < 0 {
		total = 0
	}
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}
```

Note: `time` needs to be added to the import block above (`"time"`) — included in the final file, just called out here since the step above's snippet starts mid-file conceptually.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS (all `internal/ui` tests, including the 4 new ones)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/ui/...
make lint
```
Expected: both clean. Write out this commit message and stop:

```
feat(ui): add banner color-sweep renderer for now-playing tracks

Renders the existing ASCII banner unchanged when idle; while a
track plays, sweeps each character from muted to accent color
proportional to elapsed/duration, with the title and m:ss/m:ss
readout on the line below — not yet wired into any screen's View()
(Task 8 does that).
```

---

### Task 7: `internal/ui/library.go` — type-ahead filter + Enter-on-track starts playback

**Files:**
- Modify: `internal/ui/library.go`
- Modify: `internal/ui/ui_test.go` (new tests appended)

**Interfaces:**
- Consumes: `player.Track`, `player.State` (Task 4).
- Produces: `libState` gains a `filter string` field and a `(libState).visible() []string` method; `updateLibrary` handles filter typing and clears the old `"q"`-returns-to-menu shortcut (redundant with `esc`, and `"q"` must be typeable into the filter); `libraryEnter()` at the track level returns a new `startPlaybackMsg{tracks []player.Track, index int}` (a `tea.Cmd`-carried message, following this codebase's existing `taskDoneMsg`/`progressMsg` pattern) instead of doing nothing — Task 8 wires this message to actually call `player.Load`. This task also adds a `playback player.State` field to `model` (used by `viewLibrary`'s banner call below) — Task 8 adds the rest of the playback-related `model` fields (`player`, `playerErr`) and is what actually keeps `playback` populated (via the spinner tick); until Task 8 lands, `playback` just stays at its zero value and every screen renders the plain idle banner, same as today.

**Why `"q"` is removed from Library specifically:** with a type-ahead filter, every printable character — including `q` — must reach the filter, not a shortcut. `esc`/`left`/`h` already return to the menu from the top (Artists) level via the existing `libraryBack()`, so `"q"` was redundant. This mirrors a real bug (letters silently swallowed as navigation instead of reaching a type-ahead filter) that a previous UI redesign attempt hit and fixed the same way — fixing it here now instead of later.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/ui_test.go — append these tests

func TestLibraryTypeAheadFiltersEntries(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes", "Kanye West", "Aphex Twin"}}

	next, _ := m.updateLibrary(key("k")) // 'k' must filter, not be treated as "up"
	m = next.(model)
	next, _ = m.updateLibrary(key("a"))
	m = next.(model)
	next, _ = m.updateLibrary(key("n"))
	m = next.(model)

	if m.lib.filter != "kan" {
		t.Fatalf("filter = %q, want %q", m.lib.filter, "kan")
	}
	visible := m.lib.visible()
	if len(visible) != 1 || visible[0] != "Kanye West" {
		t.Fatalf("visible() = %v, want [Kanye West]", visible)
	}
}

func TestLibraryFilterBackspaceShortens(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes"}, filter: "nuj"}

	next, _ := m.updateLibrary(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(model)
	if m.lib.filter != "nu" {
		t.Fatalf("filter after backspace = %q, want %q", m.lib.filter, "nu")
	}
}

func TestLibraryFilterResetsOnLevelChange(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes"}, cursor: 0, filter: "nuj"}

	next, _ := m.libraryEnter()
	m = next.(model)
	if m.lib.filter != "" {
		t.Errorf("filter after drilling into a new level = %q, want empty", m.lib.filter)
	}
}

func TestLibraryQIsFilterableNotAShortcut(t *testing.T) {
	m := newTestModel()
	m.screen = screenLibrary // must be set explicitly: updateLibrary alone never changes m.screen except via libraryBack(), so this proves 'q' didn't trigger that path, not just that the screen was already screenMenu to begin with
	m.lib = libState{level: 0, entries: []string{"Queen"}}

	next, _ := m.updateLibrary(key("q"))
	m = next.(model)
	if m.screen != screenLibrary {
		t.Fatalf("'q' must not exit Library while type-ahead filtering is active, got screen = %v", m.screen)
	}
	if m.lib.filter != "q" {
		t.Errorf("filter = %q, want %q", m.lib.filter, "q")
	}
}

func TestLibraryEnterOnTrackReturnsStartPlaybackCmd(t *testing.T) {
	m := newTestModel()
	m.lib = libState{
		level:   2,
		artist:  "Nujabes",
		album:   "Modal Soul",
		cursor:  1,
		entries: []string{"01 Feather.opus", "02 Reflection Eternal.opus"},
	}

	_, cmd := m.libraryEnter()
	if cmd == nil {
		t.Fatal("expected a cmd starting playback, got nil")
	}
	msg := cmd()
	sp, ok := msg.(startPlaybackMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want startPlaybackMsg", msg)
	}
	if sp.index != 1 || len(sp.tracks) != 2 {
		t.Fatalf("startPlaybackMsg = %+v, want index=1 and 2 tracks", sp)
	}
	if sp.tracks[1].Title != "02 Reflection Eternal" {
		t.Errorf("tracks[1].Title = %q, want %q (extension stripped)", sp.tracks[1].Title, "02 Reflection Eternal")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -run TestLibrary -v`
Expected: FAIL — `libState.visible`, `startPlaybackMsg` undefined; `updateLibrary` still treats letters as shortcuts.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/library.go — full replacement
package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/player"
)

// libState tracks navigation through the Artist/Album/Track hierarchy, plus
// a live type-ahead filter over the current level's entries.
type libState struct {
	level   int // 0 = artists, 1 = albums, 2 = tracks
	cursor  int
	artist  string
	album   string
	entries []string // full, unfiltered list for the current level
	filter  string
}

// visible returns entries narrowed by filter (case-insensitive substring
// match), or entries unchanged when filter is empty.
func (l libState) visible() []string {
	if l.filter == "" {
		return l.entries
	}
	lower := strings.ToLower(l.filter)
	var out []string
	for _, e := range l.entries {
		if strings.Contains(strings.ToLower(e), lower) {
			out = append(out, e)
		}
	}
	return out
}

// startPlaybackMsg asks the root model to start playing tracks[index].
type startPlaybackMsg struct {
	tracks []player.Track
	index  int
}

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true, ".webm": true, ".aac": true}

func (m model) openLibrary() (tea.Model, tea.Cmd) {
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: listDirs(m.cfg.MusicDirectory)}
	return m, nil
}

func (m model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyLeft:
		return m.libraryBack()
	case tea.KeyUp:
		if m.lib.cursor > 0 {
			m.lib.cursor--
		}
	case tea.KeyDown:
		visible := m.lib.visible()
		if m.lib.cursor < len(visible)-1 {
			m.lib.cursor++
		}
	case tea.KeyRight, tea.KeyEnter:
		return m.libraryEnter()
	case tea.KeyBackspace:
		if len(m.lib.filter) > 0 {
			r := []rune(m.lib.filter)
			m.lib.filter = string(r[:len(r)-1])
			m.lib.cursor = 0
		}
	case tea.KeyRunes:
		// Every printable rune reaches the filter unconditionally — no
		// vim-key (h/j/k/l) shortcuts here, since any of those letters can
		// legitimately start a search query (e.g. "kanye", "long"), and a
		// guard like "only when filter is empty" is always true for the
		// very first keystroke, which would silently eat it as navigation
		// instead. Real navigation already has dedicated keys above
		// (arrows, enter, esc, backspace). (An earlier draft of this
		// function DID add h/j/k/l as vim-style shortcuts guarded by
		// `filter == ""` — caught by TestLibraryTypeAheadFiltersEntries
		// during Task 7's own execution: typing "kan" produced filter
		// "an", since the leading "k" was swallowed as "up" every time,
		// the exact collision class this design was supposed to avoid.)
		m.lib.filter += msg.String()
		m.lib.cursor = 0
	}
	return m, nil
}

func (m model) libraryBack() (tea.Model, tea.Cmd) {
	switch m.lib.level {
	case 0:
		m.screen = screenMenu
	case 1:
		m.lib = libState{level: 0, entries: listDirs(m.cfg.MusicDirectory)}
	case 2:
		artistDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist)
		m.lib = libState{level: 1, artist: m.lib.artist, entries: listDirs(artistDir)}
	}
	return m, nil
}

func (m model) libraryEnter() (tea.Model, tea.Cmd) {
	visible := m.lib.visible()
	if len(visible) == 0 {
		return m, nil
	}
	if m.lib.cursor >= len(visible) {
		m.lib.cursor = len(visible) - 1
	}
	selected := visible[m.lib.cursor]

	switch m.lib.level {
	case 0:
		artistDir := filepath.Join(m.cfg.MusicDirectory, selected)
		m.lib = libState{level: 1, artist: selected, entries: listDirs(artistDir)}
	case 1:
		albumDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist, selected)
		m.lib = libState{level: 2, artist: m.lib.artist, album: selected, entries: listTracks(albumDir)}
	case 2:
		albumDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist, m.lib.album)
		index := 0
		for i, name := range m.lib.entries {
			if name == selected {
				index = i
			}
		}
		tracks := make([]player.Track, len(m.lib.entries))
		for i, name := range m.lib.entries {
			tracks[i] = player.Track{
				Path:  filepath.Join(albumDir, name),
				Title: strings.TrimSuffix(name, filepath.Ext(name)),
			}
		}
		return m, func() tea.Msg { return startPlaybackMsg{tracks: tracks, index: index} }
	}
	return m, nil
}

func (m model) viewLibrary() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")

	var crumb string
	switch m.lib.level {
	case 0:
		crumb = "Library / Artists"
	case 1:
		crumb = "Library / " + m.lib.artist
	case 2:
		crumb = "Library / " + m.lib.artist + " / " + m.lib.album
	}
	b.WriteString("  " + m.st.subtitle.Render(crumb) + "\n\n")

	visible := m.lib.visible()
	if m.lib.filter != "" {
		b.WriteString("  " + m.st.help.Render("filter: "+m.lib.filter) + "\n\n")
	}
	if len(visible) == 0 {
		b.WriteString("  " + m.st.item.Render("(no matches)") + "\n")
	}
	for i, e := range visible {
		cursor := "  "
		label := m.st.item.Render(e)
		if i == m.lib.cursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(e)
		}
		b.WriteString("  " + cursor + label + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  type to filter   enter: open   esc: back"))
	return b.String()
}

// listDirs returns the sorted names of subdirectories in dir.
func listDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// listTracks returns the sorted names of audio files in dir.
func listTracks(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
```

Also add the `playback player.State` field to the `model` struct in `internal/ui/program.go` (just the field — Task 8 adds `player`/`playerErr` and the logic that actually populates it):

```go
// Add to the model struct in internal/ui/program.go:
	playback player.State
```

This one field is enough for `viewLibrary`'s `renderBannerWithPlayback(m.st, m.theme, m.playback)` call above to compile — `renderBannerWithPlayback` already exists from Task 6, and `player.State{}`'s zero value renders the plain idle banner (per `TestRenderBannerWithPlaybackIdleMatchesPlainBanner` in Task 6), so every screen keeps looking exactly as it does today until Task 8 wires up the rest.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS (all `internal/ui` tests, including the 5 new ones)

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
go vet ./internal/ui/...
make lint
```
Expected: both clean. Write out this commit message and stop:

```
feat(ui): add Library type-ahead search and wire track selection to playback

Typing while browsing any level (Artists/Albums/Tracks) filters the
list live; navigation switches from string key-matching to
tea.KeyMsg.Type so every printable rune — including q, previously a
"back to menu" shortcut — reaches the filter instead of being
swallowed as a shortcut. Selecting a track now returns a
startPlaybackMsg carrying the full album as the queue (not yet
consumed anywhere — Task 8 wires it to the player).
```

---

### Task 8: `internal/ui/program.go` — global playback key routing, banner wiring, player lifecycle

**Files:**
- Modify: `internal/ui/program.go`
- Modify: `internal/ui/library.go` (add the `playerErr`/`playback.LastError` display lines to `viewLibrary`, which already renders through `renderBannerWithPlayback` as of Task 7)
- Modify: `internal/ui/settings.go`, `internal/ui/menu.go`, `internal/ui/picker.go`, `internal/ui/forms.go`, `internal/ui/theme_picker.go`, `internal/ui/log_view.go` (each `view*()` function's banner line)
- Modify: `internal/ui/ui_test.go` (new tests appended)

**Interfaces:**
- Consumes: `player.Available`, `player.New`, `(*player.Player).{State,TogglePause,Stop,SeekBy,Next,Prev,Close,Load}` (Task 4); `startPlaybackMsg` (Task 7); `renderBannerWithPlayback` (Task 6); `model.playback` (already added to `model` in Task 7).
- Produces: `model.player *player.Player` (nil until first successful `Load`), `model.playerErr string` (a one-shot inline error line, cleared on the next keypress); wires the spinner tick to actually refresh `model.playback` every tick (the field itself already exists from Task 7, sitting at its zero value until this task).

This is the integration task — it has no isolated unit-testable slice of its own beyond the key-routing and message-handling logic below, so it follows the existing `ui_test.go` pattern of driving `model.Update`/`handleKey` directly rather than a fresh red/green cycle per sub-piece.

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/ui_test.go — append these tests

func TestGlobalPlaybackKeysDoNothingWithoutAPlayer(t *testing.T) {
	m := newTestModel()
	m.screen = screenMenu
	// No player loaded yet — space/n/p/s/arrows must not panic.
	for _, k := range []string{" ", "n", "p", "s"} {
		next, _ := m.handleKey(key(k))
		m = next.(model)
	}
	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = next.(model)
	_ = next
}

func TestSpacebarDoesNotTogglePauseOnLibraryScreen(t *testing.T) {
	m := newTestModel()
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: []string{"Nujabes"}}

	next, _ := m.handleKey(key(" "))
	m = next.(model)
	// " " must have been treated as a filter character, not a global pause.
	if m.lib.filter != " " {
		t.Errorf("filter = %q, want a single space (global pause must not intercept it here)", m.lib.filter)
	}
}

func TestSpacebarDoesNotTogglePauseOnInputScreen(t *testing.T) {
	m := newTestModel()
	next, _ := m.openInput("search", "Enter query")
	m = next.(model)

	valueBefore := m.input.Value()
	next, _ = m.updateInput(key(" "))
	m = next.(model)
	if m.input.Value() == valueBefore {
		t.Error("space must reach the text input on the Input screen, not be swallowed as global pause")
	}
}

func TestStartPlaybackMsgWithoutMpvSetsPlayerErr(t *testing.T) {
	if player.Available() {
		t.Skip("mpv is installed on this machine; this test only covers the not-installed path")
	}
	m := newTestModel()
	next, _ := m.Update(startPlaybackMsg{tracks: []player.Track{{Path: "/x.opus", Title: "X"}}, index: 0})
	m = next.(model)
	if m.playerErr == "" {
		t.Error("expected playerErr to be set when mpv is not installed")
	}
	if m.player != nil {
		t.Error("expected model.player to stay nil when mpv is not installed")
	}
}

func TestPlayerErrClearsOnNextKeyPress(t *testing.T) {
	m := newTestModel()
	m.playerErr = "mpv not found"
	next, _ := m.handleKey(key("x"))
	m = next.(model)
	if m.playerErr != "" {
		t.Error("expected playerErr to clear on the next keypress")
	}
}

func TestPlaybackLastErrorRendersOnMenuAndLibrary(t *testing.T) {
	m := newTestModel()
	m.playback = player.State{LastError: "playback error, skipping track"}

	if !strings.Contains(m.viewMenu(), "playback error, skipping track") {
		t.Error("expected LastError to render on the menu screen")
	}

	next, _ := m.openLibrary()
	m = next.(model)
	m.playback = player.State{LastError: "playback error, skipping track"}
	if !strings.Contains(m.viewLibrary(), "playback error, skipping track") {
		t.Error("expected LastError to render on the library screen")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -run 'TestGlobalPlaybackKeys|TestSpacebar|TestStartPlaybackMsg|TestPlayerErr' -v`
Expected: FAIL — `model.player`/`model.playback`/`model.playerErr` undefined, `handleKey` doesn't scope global keys yet.

- [ ] **Step 3: Write the implementation**

```go
// internal/ui/program.go — additions and modifications

// Add to imports:
//   "time"
//   "github.com/EnJulian/shadowbox/internal/player"

// Add to the model struct (playback player.State already added in Task 7):
	player    *player.Player
	playerErr string

// Add a helper near handleKey:

// screenCapturesText reports whether the given screen consumes free-text
// keystrokes (a live filter or text field), and must not have global
// playback keys (space/n/p/s/arrows) stolen out from under it.
func screenCapturesText(s screen) bool {
	switch s {
	case screenLibrary, screenInput, screenSettingEdit:
		return true
	}
	return false
}

// Replace handleKey's body with:

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.playerErr = ""

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.screen == screenMenu && msg.String() == "q" {
		return m, tea.Quit
	}

	if !screenCapturesText(m.screen) {
		if next, cmd, handled := m.handlePlaybackKey(msg); handled {
			return next, cmd
		}
	}

	switch m.screen {
	case screenMenu:
		return m.updateMenu(msg)
	case screenInput:
		return m.updateInput(msg)
	case screenSettings:
		return m.updateSettings(msg)
	case screenSettingEdit:
		return m.updateSettingEdit(msg)
	case screenThemePicker:
		return m.updateThemePicker(msg)
	case screenLibrary:
		return m.updateLibrary(msg)
	case screenDownloadLog:
		return m.updateDownloadLog(msg)
	case screenPicker:
		return m.updatePicker(msg)
	case screenResult:
		m.screen = screenMenu
		return m, nil
	case screenRunning:
		return m, nil
	}
	return m, nil
}

// handlePlaybackKey handles the global playback keys. handled is false for
// any other key, so the caller falls through to the screen's own handling.
func (m model) handlePlaybackKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.player == nil {
		switch msg.String() {
		case " ", "n", "p", "s":
			return m, nil, true // consume the key, but there's nothing to control yet
		}
		if msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight {
			return m, nil, true
		}
		return m, nil, false
	}

	switch msg.String() {
	case " ":
		_ = m.player.TogglePause()
		return m, nil, true
	case "n":
		_ = m.player.Next()
		return m, nil, true
	case "p":
		_ = m.player.Prev()
		return m, nil, true
	case "s":
		_ = m.player.Stop()
		return m, nil, true
	}
	switch msg.Type {
	case tea.KeyLeft:
		_ = m.player.SeekBy(-10 * time.Second)
		return m, nil, true
	case tea.KeyRight:
		_ = m.player.SeekBy(10 * time.Second)
		return m, nil, true
	}
	return m, nil, false
}

// Add a new case to Update's message switch, alongside taskDoneMsg/progressMsg:

	case startPlaybackMsg:
		if m.player == nil {
			if !player.Available() {
				m.playerErr = "mpv not found — install it to enable playback"
				return m, nil
			}
			p, err := player.New()
			if err != nil {
				m.playerErr = "failed to start mpv: " + err.Error()
				return m, nil
			}
			m.player = p
		}
		if err := m.player.Load(msg.tracks, msg.index); err != nil {
			m.playerErr = "failed to play: " + err.Error()
		}
		return m, nil

// Add a case to the spinner.TickMsg branch (which already fires continuously)
// so the banner sweep updates every tick without a new ticker:

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.player != nil {
			m.playback = m.player.State()
		}
		return m, cmd

// In runProgram, clean up the player process on exit:

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if fm, ok := final.(model); ok && fm.player != nil {
		_ = fm.player.Close()
	}
	return err
```

`internal/ui/library.go`'s `viewLibrary` already renders through `renderBannerWithPlayback` as of Task 7 — this task only adds its `playerErr`/`playback.LastError` display lines (below). Apply the one-line banner swap — replacing `b.WriteString(m.st.title.Render(banner))` with `b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))` — to every *other* screen's `view*()`:
- `internal/ui/menu.go` (`viewMenu`, shown fully below along with its error lines)
- `internal/ui/settings.go` (`viewSettings`, `viewSettingEdit`)
- `internal/ui/theme_picker.go` (`viewThemePicker`)
- `internal/ui/log_view.go` (`viewDownloadLog`)
- `internal/ui/picker.go` (`viewPicker`)
- `internal/ui/forms.go` (`viewInput`, `viewRunning`, `viewResult`, or wherever that file's `view*()` functions render the banner — check the file for the exact function names and banner call site(s) before editing, since this plan was written against the file list as of Task 7, not a fresh read of `forms.go`)

`viewMenu` (currently in `internal/ui/menu.go`) additionally gets an error line — the one screen a failed or interrupted playback attempt is most likely to be viewed from, alongside `viewLibrary` (edited in this task's `library.go` change). Full replacement for `viewMenu`:

```go
func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString(m.st.subtitle.Render("  Music acquisition console"))
	b.WriteString("\n\n")

	for i, item := range mainMenu {
		cursor := "  "
		line := m.st.item.Render(item)
		if i == m.menuCursor {
			cursor = m.st.accent.Render("> ")
			line = m.st.selected.Render(item)
		}
		b.WriteString("  " + cursor + line + "\n")
	}

	if m.playerErr != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playerErr) + "\n")
	}
	if m.playback.LastError != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playback.LastError) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: select   q: quit"))
	return b.String()
}
```

And in `viewLibrary` (`internal/ui/library.go`, from Task 7), insert the same two conditional blocks right before its final `b.WriteString("\n")` / help-line pair:

```go
	if m.playerErr != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playerErr) + "\n")
	}
	if m.playback.LastError != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playback.LastError) + "\n")
	}
```

`m.playerErr` is cleared on the next keypress (already handled in `handleKey`, Task 8's other change). `m.playback.LastError` is not cleared by a keypress — it naturally goes away once the next successful `Load` runs (auto-advance already resets it via `State{}`'s zero value, per Task 4), so no additional clearing logic is needed for it here.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS (all `internal/ui` tests)

Run: `go build ./...`
Expected: no errors — this confirms every `view*()` call site was updated consistently (a leftover `m.st.title.Render(banner)` call site is not a build error by itself, but a missed one means that screen won't show the sweep; grep for `title.Render(banner)` after editing and confirm zero remaining matches outside `bannersweep.go` itself).

- [ ] **Step 5: Verify and stop (do not commit)**

```bash
grep -rn "st.title.Render(banner)" internal/ui/*.go
```
Expected: no matches (confirms every screen was migrated).

```bash
go vet ./...
make lint
make test
```
Expected: all clean. Write out this commit message and stop:

```
feat(ui): wire playback into the root model and every screen's banner

Global playback keys (space/n/p/s/left/right) work from any screen
except Library/Input/Settings-edit, which need those keys for
text/filter entry instead. startPlaybackMsg lazily starts mpv (with
a clear inline error if it's not installed) and loads the selected
track's album as the queue. Every screen now renders the banner
through renderBannerWithPlayback, so the color-sweep and
now-playing line appear identically everywhere, driven by the
existing spinner tick rather than a new ticker.
```

---

### Task 9: Documentation — mpv as a new dependency

**Files:**
- Modify: `README.md`
- Modify: `docs/INSTALL_FROM_SOURCE.md`
- Modify: `docs/INSTALL_WINDOWS.md`

**Interfaces:** none — documentation only.

- [ ] **Step 1: Update `README.md`**

Find the existing line documenting required tools (`Needs yt-dlp and ffmpeg on your PATH (aria2 optional)`) and update it to:

```
Needs `yt-dlp` and `ffmpeg` on your PATH (`aria2` optional, `mpv` optional — required only for in-app playback).
```

Find the credits/external-tools paragraph near the license section (mentions yt-dlp, FFmpeg, aria2) and add mpv:

```
MIT — see [LICENSE](LICENSE). Shadowbox invokes [yt-dlp](https://github.com/yt-dlp/yt-dlp),
[FFmpeg](https://ffmpeg.org), [aria2](https://aria2.github.io), and [mpv](https://mpv.io)
as external tools; it does not bundle or redistribute them.
```

- [ ] **Step 2: Update `docs/INSTALL_FROM_SOURCE.md`**

Add `mpv` to whatever list of `apt`/`brew`/package-manager install commands already covers `yt-dlp`/`ffmpeg`, noting it's optional and playback-only — read the file first to match its existing per-platform structure exactly rather than guessing the format.

- [ ] **Step 3: Update `docs/INSTALL_WINDOWS.md`**

Add `mpv` to the Windows install instructions the same way — read the file first (it likely references the Scoop bucket, which bundles `ffmpeg`/`yt-dlp` today per the README) and note that `mpv` is a separate, optional install (`scoop install mpv` if it's in the standard Scoop bucket, or a direct link to https://mpv.io/installation/ otherwise — verify which before writing the instruction).

- [ ] **Step 4: Verify and stop (do not commit)**

No automated check for documentation; proofread the three files for accuracy against what Tasks 1-8 actually built. Write out this commit message and stop:

```
docs: document mpv as an optional playback dependency

Adds mpv to the README's tool list, credits, and the source/Windows
install guides, consistent with how aria2 is already documented as
optional.
```

---

## Testing Summary

| Layer | Tests |
|-------|-------|
| `internal/player` (IPC) | Command/reply correlation, error replies, timeouts, event demuxing — all via `net.Pipe()`, no real mpv |
| `internal/player` (socket) | Unix dial success/timeout/uniqueness (Windows: cross-compile only) |
| `internal/player` (Player) | Load, State polling, pause toggle, next/prev bounds, stop, seek clamping, auto-advance on eof, no-advance on non-eof end-file — all via a fake `mpvConn` |
| `internal/download` | mpv listed as optional in `Dependencies()` |
| `internal/ui` (banner) | Idle vs. playing render, now-playing line content, sweep changes with progress, zero-duration safety |
| `internal/ui` (library) | Type-ahead filter narrows/resets/backspaces, `q` is filterable not a shortcut, track selection produces the right queue |
| `internal/ui` (routing) | Global keys inert without a player, scoped away from Library/Input's text capture, `startPlaybackMsg` error path when mpv is missing, error line clears on next keypress |

All changes must pass `make lint` and `make test` (per Global Constraints). Windows named-pipe support (Task 3) is cross-compile-verified only — flagged as an accepted, real gap in this project's coverage, not an oversight.

## Not Covered By This Plan

Per the design spec's Out of Scope section: in-app volume control, a dedicated "Now Playing" screen, manual queue reordering/playlists/shuffle/repeat, crossfade or gapless-playback tuning, any visualizer beyond the banner sweep, and volume normalization/ReplayGain.
