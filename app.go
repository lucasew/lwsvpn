package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	socks5 "github.com/armon/go-socks5"
	"github.com/lucasew/wsvpn/pkg/errors"
	"golang.org/x/net/websocket"
)

var (
	err      error
	port     int
	secret   string
	logfile  *bytes.Buffer
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
	fmt.Fprintf(logfile, "spawning: %s %+v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(logfile, "%s: %s\n", name, err.Error())
	}
}

func RootHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, err := fmt.Fprintln(w, "Hello, world")
	errors.ReportError(err)
}

func LogHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, err := io.Copy(w, logfile)
	errors.ReportError(err)
}

func main() {
	go SpawnProgram("wstunnel", "--server", "ws://0.0.0.0:3000")
	go SpawnProgram("rclone", "serve", "webdav", "--addr", ":9999", "davsrv:/", "--config", "./rclone.conf")
	http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)
	http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(func(conn *websocket.Conn) {
		err := socksSrv.ServeConn(conn)
		errors.ReportError(err)
	}))
	http.HandleFunc("/", RootHTTPHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}
