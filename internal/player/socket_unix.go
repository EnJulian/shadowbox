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
