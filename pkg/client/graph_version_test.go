package client

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

// graphVersionServer records the graph version echoed by the client and replies
// with the versions given, one per request.
type graphVersionServer struct {
	*httptest.Server
	mutex    sync.Mutex
	received []string
	replies  []string
	calls    int
}

func newGraphVersionServer(replies ...string) *graphVersionServer {
	server := &graphVersionServer{replies: replies}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.mutex.Lock()
		server.received = append(server.received, r.Header.Get(GraphVersionHeader))
		var reply string
		if server.calls < len(server.replies) {
			reply = server.replies[server.calls]
		}
		server.calls++
		server.mutex.Unlock()

		if reply != "" {
			w.Header().Set(GraphVersionHeader, reply)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	return server
}

func (s *graphVersionServer) requestHeaders() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]string(nil), s.received...)
}

// makeGraphVersionClient builds a client pointed at the test server. Make does a
// catalog call at startup, so the counting starts from a clean server.
func makeGraphVersionClient(t *testing.T, server *graphVersionServer) *Client {
	t.Helper()
	client, err := Make(APIParameter{
		APIKey:  "aToken",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("could not build client: %s", err)
	}
	server.mutex.Lock()
	server.received = nil
	server.calls = 0
	server.mutex.Unlock()
	// Reset in place: the response middleware closed over this tracker, so
	// swapping in a new one would leave the test reading a different counter
	// from the one the client updates.
	client.graphVersion.version.Store(0)
	return client
}

func (client *Client) getIgnoringError(t *testing.T) {
	t.Helper()
	_, err := client.client.R().Get(client.baseURL + "/whatever")
	if err != nil {
		t.Fatalf("request failed: %s", err)
	}
}

func TestGraphVersionIsCapturedFromResponse(t *testing.T) {
	server := newGraphVersionServer("42")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)

	if got := client.graphVersion.current(); got != 42 {
		t.Errorf("expected tracked version 42, got %d", got)
	}
}

func TestGraphVersionIsEchoedOnNextRequest(t *testing.T) {
	server := newGraphVersionServer("42", "")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)
	client.getIgnoringError(t)

	headers := server.requestHeaders()
	if len(headers) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(headers))
	}
	if headers[0] != "" {
		t.Errorf("first request should carry no graph version, got %q", headers[0])
	}
	if headers[1] != "42" {
		t.Errorf("second request should echo 42, got %q", headers[1])
	}
}

func TestGraphVersionDoesNotRegressOnLowerVersion(t *testing.T) {
	server := newGraphVersionServer("42", "7", "")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)
	client.getIgnoringError(t)
	client.getIgnoringError(t)

	if got := client.graphVersion.current(); got != 42 {
		t.Errorf("expected tracked version to stay at 42, got %d", got)
	}
	headers := server.requestHeaders()
	if headers[2] != "42" {
		t.Errorf("third request should still echo 42, got %q", headers[2])
	}
}

func TestGraphVersionAbsentFromResponseSendsNoHeader(t *testing.T) {
	server := newGraphVersionServer("", "")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)
	client.getIgnoringError(t)

	for i, header := range server.requestHeaders() {
		if header != "" {
			t.Errorf("request %d should carry no graph version, got %q", i, header)
		}
	}
	if got := client.graphVersion.current(); got != 0 {
		t.Errorf("expected no tracked version, got %d", got)
	}
}

func TestGraphVersionIgnoresUnparsableHeader(t *testing.T) {
	server := newGraphVersionServer("not-a-number", "")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)
	client.getIgnoringError(t)

	if got := client.graphVersion.current(); got != 0 {
		t.Errorf("expected unparsable version to be ignored, got %d", got)
	}
	if header := server.requestHeaders()[1]; header != "" {
		t.Errorf("second request should carry no graph version, got %q", header)
	}
}

// Applies run concurrently on a shared client, so the tracker must keep the max
// under parallel updates.
func TestGraphVersionTrackerIsThreadSafe(t *testing.T) {
	tracker := &graphVersionTracker{}
	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(v int64) {
			defer wg.Done()
			tracker.observe(v)
		}(int64(i))
	}
	wg.Wait()

	if got := tracker.current(); got != 100 {
		t.Errorf("expected max version 100, got %d", got)
	}
}

func TestGraphVersionSurvivesAuthHeaderSetup(t *testing.T) {
	server := newGraphVersionServer("42", "")
	defer server.Close()
	client := makeGraphVersionClient(t, server)

	client.getIgnoringError(t)
	// SetAPIKey reassigns the resty client field; the echo must still be installed.
	client.SetAPIKey("anotherToken")
	client.getIgnoringError(t)

	if header := server.requestHeaders()[1]; header != strconv.FormatInt(42, 10) {
		t.Errorf("expected graph version to survive auth setup, got %q", header)
	}
}
