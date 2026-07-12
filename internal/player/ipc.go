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
// both closed and eventsCh, so Player's watchEvents range loop terminates
// and can tell the difference between "waiting" and "gone."
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
	// wait below.
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
// finishing). Closed once the underlying connection closes.
func (c *ipcClient) Events() <-chan ipcEvent {
	return c.eventsCh
}

// close shuts down the underlying connection.
func (c *ipcClient) close() error {
	return c.conn.Close()
}
