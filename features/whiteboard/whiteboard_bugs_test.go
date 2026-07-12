package whiteboard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/calionauta/gogogo-fullstack-template/internal/collab"
)

// TestWhiteboard_PresenceToleratesStringCoords is the regression guard for
// the "POST /api/whiteboard/<doc>/presence 400 (Bad Request)" bug.
//
// Root cause: the client posted cursor coordinates as JSON strings
// (x: "0.5" via .toFixed), but PresenceMsg.X/Y are float64, so
// json.Unmarshal failed and the handler returned 400. That 400 also broke
// remote cursors (BUG: cursors never arrived). The fix makes the server
// tolerant of numeric strings AND the client now sends numbers. This test
// proves both directions: a string-coords payload is accepted (200) and a
// number-coords payload is accepted (200). The server must never 400 a
// well-formed cursor just because a client (or a fork) stringified a number.
func TestWhiteboard_PresenceToleratesStringCoords(t *testing.T) {
	baseURL, _, cleanup := webFixture(t)
	defer cleanup()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	client := &http.Client{
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	login(t, client, baseURL)

	docID := "doc-presstr-" + time.Now().Format("150405.000")

	// String coords — the exact payload that previously 400'd.
	strBody := []byte(`{"type":"cursor","doc":"` + docID + `","user":"u-str","x":"0.1234","y":"0.4321","ts":1}`)
	resp, err := postWithClientID(context.Background(), client, baseURL+"/api/whiteboard/"+docID+"/presence", "wb-str", strBody)
	if err != nil {
		t.Fatalf("presence (string) POST: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("presence (string coords) should be 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Number coords — the corrected client payload.
	numBody, _ := json.Marshal(collab.PresenceMsg{Type: "cursor", Doc: docID, User: "u-num", X: 0.5, Y: 0.5, TS: 1})
	resp2, err := postWithClientID(context.Background(), client, baseURL+"/api/whiteboard/"+docID+"/presence", "wb-num", numBody)
	if err != nil {
		t.Fatalf("presence (number) POST: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		resp2.Body.Close()
		t.Fatalf("presence (number coords) should be 200, got %d", resp2.StatusCode)
	}
	resp2.Body.Close()
}

// TestWhiteboard_PeerCountAuthoritative is the regression guard for the
// "online count is wrong / inconsistent across tabs (one shows 3, another
// shows 1, but only 2 tabs)" bug.
//
// Root cause: each tab computed its own count from client-side join/leave
// increments, which drifted when a leave was missed or a tab reconnected
// with a fresh client id. The fix makes the server the authority: on every
// join and leave it broadcasts a "count" event carrying the FULL
// peer set to ALL clients, and every tab renders the count directly from
// that event. This test asserts two distinct SSE connections on the same
// doc both receive a "count" event whose peer set contains both of them
// (size 2) — so they always agree on "2 online".
func TestWhiteboard_PeerCountAuthoritative(t *testing.T) {
	baseURL, _, cleanup := webFixture(t)
	defer cleanup()

	clientA := newWBClient(t)
	clientB := newWBClient(t)
	login(t, clientA, baseURL)
	login(t, clientB, baseURL)

	docID := "doc-count-" + time.Now().Format("150405.000")
	streamA := openWBStream(t, clientA, baseURL, docID, "wbA")
	streamB := openWBStream(t, clientB, baseURL, docID, "wbB")
	defer streamA.close()
	defer streamB.close()
	time.Sleep(200 * time.Millisecond)

	peersA := countPeersFromEvents(streamA.drain(400 * time.Millisecond))
	peersB := countPeersFromEvents(streamB.drain(400 * time.Millisecond))

	if len(peersA) != 2 {
		t.Fatalf("clientA expected authoritative count of 2 peers, got %v", peersA)
	}
	if len(peersB) != 2 {
		t.Fatalf("clientB expected authoritative count of 2 peers, got %v", peersB)
	}
	// Each side must see the other's client id in the server's set.
	if !contains(peersA, "wbB") || !contains(peersB, "wbA") {
		t.Fatalf("peer sets disagree: A=%v B=%v", peersA, peersB)
	}
}

// TestWhiteboard_SingleOnlineLabel is the regression guard for the
// "badge shows '1 online' and then ANOTHER 'online' to the right" bug.
// Asserts the board page renders exactly ONE "online" label.
func TestWhiteboard_SingleOnlineLabel(t *testing.T) {
	baseURL, _, cleanup := webFixture(t)
	defer cleanup()
	client := newWBClient(t)
	login(t, client, baseURL)

	docID := "doc-online-" + time.Now().Format("150405.000")
	resp, err := client.Do(mustReq(t, http.MethodGet, baseURL+"/whiteboard/"+docID))
	if err != nil {
		t.Fatalf("GET /whiteboard/%s: %v", docID, err)
	}
	body := readAll(resp)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("board page status = %d", resp.StatusCode)
	}
	n := strings.Count(body, ">online<")
	if n != 1 {
		t.Fatalf("expected exactly 1 'online' label on the board page, got %d", n)
	}
}

// TestWhiteboard_ThemeToggleWired is the regression guard for the
// "dark/light toggle does nothing on the whiteboard page" bug. The
// whiteboard page does NOT load Datastar, so the toggle must work purely
// via theme.js binding the .theme-toggle button. Asserts the board page
// renders the toggle button and references theme.js.
func TestWhiteboard_ThemeToggleWired(t *testing.T) {
	baseURL, _, cleanup := webFixture(t)
	defer cleanup()
	client := newWBClient(t)
	login(t, client, baseURL)

	docID := "doc-theme-" + time.Now().Format("150405.000")
	resp, err := client.Do(mustReq(t, http.MethodGet, baseURL+"/whiteboard/"+docID))
	if err != nil {
		t.Fatalf("GET /whiteboard/%s: %v", docID, err)
	}
	body := readAll(resp)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("board page status = %d", resp.StatusCode)
	}
	if !strings.Contains(body, "theme-toggle") {
		t.Fatalf("board page missing .theme-toggle button (dark/light toggle would be inert)")
	}
	if !strings.Contains(body, "/static/theme.js") {
		t.Fatalf("board page does not load theme.js (toggle binding never runs)")
	}
}

// --- extra helpers (distinct from web_test.go) ---

func newWBClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	return &http.Client{
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func countPeersFromEvents(events []string) []string {
	var latest []string
	for _, ev := range events {
		raw := strings.TrimPrefix(strings.TrimSpace(ev), "data: ")
		var msg collab.PresenceMsg
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue
		}
		if msg.Type == "count" {
			latest = msg.Peers // keep the most recent authoritative count
		}
	}
	return latest
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
