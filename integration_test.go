package main

import (
    "net/http/httptest"
    "testing"
    "time"

    "golang.org/x/net/websocket"
    "github.com/hashicorp/yamux"
)

func TestYamuxHandshake(t *testing.T) {
    // Start a test server with the extracted WebSocketHandler
    // WebSocketHandler uses socksSrv which is initialized in init()
    // init() requires PORT and SECRET env vars.
    // We assume they are provided when running tests.

    // However, if socksSrv is nil (e.g. init panicked or didn't run properly before), this will panic.
    // But init() runs on package initialization.

    if socksSrv == nil {
        t.Skip("socksSrv not initialized, skipping test. Ensure PORT and SECRET are set.")
    }

    ts := httptest.NewServer(websocket.Handler(WebSocketHandler))
    defer ts.Close()

    // Connect client
    wsURL := "ws" + ts.URL[4:]
    ws, err := websocket.Dial(wsURL, "", "http://localhost/")
    if err != nil {
        t.Fatalf("Failed to connect to websocket: %v", err)
    }
    defer ws.Close()

    // Establish Yamux session
    session, err := yamux.Client(ws, nil)
    if err != nil {
        t.Fatalf("Failed to create yamux client: %v", err)
    }
    defer session.Close()

    // Open a stream
    stream, err := session.Open()
    if err != nil {
        t.Fatalf("Failed to open yamux stream: %v", err)
    }
    defer stream.Close()

    // Write something to the stream
    // Since socks5 server expects handshake, we can try to send a SOCKS5 greeting.
    // version 5, 1 auth method (0x00 - no auth)
    greeting := []byte{0x05, 0x01, 0x00}
    _, err = stream.Write(greeting)
    if err != nil {
        t.Fatalf("Failed to write to stream: %v", err)
    }

    // Read response
    // Server should respond with version 5, method 0
    resp := make([]byte, 2)
    // Set deadline
    stream.SetReadDeadline(time.Now().Add(5 * time.Second))
    _, err = stream.Read(resp)
    if err != nil {
        t.Fatalf("Failed to read from stream: %v", err)
    }

    if resp[0] != 0x05 || resp[1] != 0x00 {
        t.Fatalf("Unexpected socks5 response: %v", resp)
    }
}
