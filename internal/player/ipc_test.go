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
