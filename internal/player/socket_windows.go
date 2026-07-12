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
