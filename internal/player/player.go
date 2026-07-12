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
