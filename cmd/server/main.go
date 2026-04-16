package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
    "errors"

	pkgerrors "github.com/lucasew/wsvpn/pkg/errors"

	socks5 "github.com/armon/go-socks5"
	"golang.org/x/net/websocket"
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
        pkgerrors.ReportFatal(err, "failed to parse PORT environment variable")
    }
    secret = os.Getenv("SECRET")
    if secret == "" {
        pkgerrors.ReportFatal(errors.New("SECRET is not defined"), "missing SECRET")
    }
    logfile = bytes.NewBuffer([]byte{})
    log.SetOutput(logfile)

    socksSrv, err = socks5.New(&socks5.Config{})
    if err != nil {
        pkgerrors.ReportFatal(err, "failed to initialize socks5 server")
    }
}

func SpawnProgram(name string, args ...string) {
    fmt.Fprintf(logfile, fmt.Sprintf("spawning: %s %+v\n", name, args))
    cmd := exec.Command(name, args...)
    cmd.Stdout = logfile
    cmd.Stderr = logfile
    cmd.Env = os.Environ()
    err := cmd.Run()
    if err != nil {
        pkgerrors.ReportError(err, fmt.Sprintf("failed to spawn program %s", name))
        fmt.Fprintf(logfile, fmt.Sprintf("%s: %s\n", name, err.Error()))
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

func main() {
    go SpawnProgram("wstunnel", "--server", "ws://0.0.0.0:3000")
    go SpawnProgram("rclone","serve", "webdav", "--addr", ":9999", "davsrv:/", "--config", "./rclone.conf")
    http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)

    http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(func(conn *websocket.Conn) {
        if err := socksSrv.ServeConn(conn); err != nil {
            pkgerrors.ReportError(err, "failed to serve socks5 connection")
        }
    }))
    http.HandleFunc("/", RootHTTPHandler)
    err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
    if err != nil {
        pkgerrors.ReportFatal(err, "failed to listen and serve")
    }
}

