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

