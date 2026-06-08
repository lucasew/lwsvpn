/*
Package main implements the server-side component of the application.
It is responsible for orchestrating background daemons (such as wstunnel and rclone)
and setting up the primary HTTP server. The server exposes a standard HTTP health check,
a log reading endpoint, and an upgraded websocket connection that acts as a SOCKS5 proxy.
*/
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
	"golang.org/x/net/websocket"
    socks5 "github.com/armon/go-socks5"
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

/*
init reads required environment variables (PORT and SECRET) and initializes
the global application state, including the shared log buffer and SOCKS5 server instance.
If required variables are missing or invalid, it immediately panics to prevent
the server from starting in an incomplete state.
*/
func init() {
    port, err = strconv.Atoi(os.Getenv("PORT"))
    if err != nil {
        panic(err)
    }
    secret = os.Getenv("SECRET")
    if secret == "" {
        panic("SECRET is not defined")
    }
    logfile = bytes.NewBuffer([]byte{})
    log.SetOutput(logfile)

    socksSrv, err = socks5.New(&socks5.Config{})
    if err != nil {
        panic(err.Error())
    }
}

/*
SpawnProgram is a helper that executes a given external command as a background process.
It binds the command's standard output and standard error directly to the shared
application logfile buffer, allowing external daemon logs to be surfaced via the HTTP log endpoint.
*/
func SpawnProgram(name string, args ...string) {
    fmt.Fprintf(logfile, fmt.Sprintf("spawning: %s %+v", name, args))
    cmd := exec.Command(name, args...)
    cmd.Stdout = logfile
    cmd.Stderr = logfile
    cmd.Env = os.Environ()
    err := cmd.Run()
    if err != nil {
        fmt.Fprintf(logfile, fmt.Sprintf("%s: %s", name, err.Error()))
    }
}

/*
RootHTTPHandler serves as a basic healthcheck endpoint at the root path ("/").
It always returns a 200 OK with a simple text greeting.
*/
func RootHTTPHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    fmt.Fprintln(w, "Hello, world")
}

/*
LogHTTPHandler exposes the contents of the in-memory log buffer to HTTP clients.
This provides a way to inspect both the main server logs and the stdout/stderr
of any daemons spawned via SpawnProgram.
*/
func LogHTTPHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    io.Copy(w, logfile)
}

/*
main coordinates the startup sequence:
1. Spawns background daemons (`wstunnel` and `rclone`) concurrently.
2. Registers the log endpoint and the websocket endpoint (protected by the SECRET).
3. Connects the websocket endpoint to the initialized SOCKS5 server.
4. Starts listening for incoming HTTP connections on the configured port.
*/
func main() {
    go SpawnProgram("wstunnel", "--server", "ws://0.0.0.0:3000")
    go SpawnProgram("rclone","serve", "webdav", "--addr", ":9999", "davsrv:/", "--config", "./rclone.conf")
    http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)
    if err != nil {
        panic(err)
    }
    http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(func(conn *websocket.Conn) {
        socksSrv.ServeConn(conn)
    }))
    http.HandleFunc("/", RootHTTPHandler)
    err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
    if err != nil {
        panic(err)
    }
}

