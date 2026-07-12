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
	// "last command" and mask what TogglePause() actually sent.
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
