package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
    "net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"golang.org/x/net/websocket"
    socks5 "github.com/armon/go-socks5"
    "github.com/lucasew/wsvpn/pkg/errors"
    "github.com/lucasew/wsvpn/pkg/errdefs"
    "github.com/hashicorp/yamux"
)

const (
    wsHttpPort = 3000
)

var (
    err error
    port int
    secret string
    ctx context.Context
    logfile *bytes.Buffer
    socksSrv *socks5.Server
)

func init() {
    port, err = strconv.Atoi(os.Getenv("PORT"))
    if err != nil {
        errors.ReportError(err, "Failed to parse PORT")
        panic(err)
    }
    secret = os.Getenv("SECRET")
    if secret == "" {
        err = fmt.Errorf("SECRET %w", errdefs.ErrNotDefined)
        errors.ReportError(err, "SECRET missing")
        panic(err)
    }
    logfile = bytes.NewBuffer([]byte{})
    log.SetOutput(logfile)

    socksSrv, err = socks5.New(&socks5.Config{})
    if err != nil {
        errors.ReportError(err, "Failed to create socks server")
        panic(err.Error())
    }
}

func SpawnProgram(name string, args ...string) {
    fmt.Fprintf(logfile, fmt.Sprintf("spawning: %s %+v", name, args))
    cmd := exec.Command(name, args...)
    cmd.Stdout = logfile
    cmd.Stderr = logfile
    cmd.Env = os.Environ()
    err := cmd.Run()
    if err != nil {
        fmt.Fprintf(logfile, fmt.Sprintf("%s: %s", name, err.Error()))
        errors.ReportError(err, fmt.Sprintf("Failed to run %s", name))
    }
}

func RootHTTPHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    fmt.Fprintln(w, "Hello, world")
}

func LogHTTPHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    io.Copy(w, logfile)
}

func WebSocketHandler(conn *websocket.Conn) {
    session, err := yamux.Server(conn, nil)
    if err != nil {
        errors.ReportError(err, "Failed to create yamux server")
        return
    }
    defer session.Close()

    for {
        stream, err := session.Accept()
        if err != nil {
            // If the session is closed, Accept returns error.
            // We just log it and exit the handler (which closes the websocket connection likely, or it was already closed).
            errors.ReportError(err, "Failed to accept yamux stream")
            break
        }
        go func(s net.Conn) {
            defer s.Close()
            if err := socksSrv.ServeConn(s); err != nil {
                errors.ReportError(err, "Socks5 ServeConn error")
            }
        }(stream)
    }
}

func main() {
    defer errors.ReportPanic()
    go SpawnProgram("wstunnel", "--server", "ws://0.0.0.0:3000")
    go SpawnProgram("rclone","serve", "webdav", "--addr", ":9999", "davsrv:/", "--config", "./rclone.conf")
    http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)

    http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(WebSocketHandler))
    http.HandleFunc("/", RootHTTPHandler)
    err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
    if err != nil {
        errors.ReportError(err, "ListenAndServe failed")
        panic(err)
    }
}
