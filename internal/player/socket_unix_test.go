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
