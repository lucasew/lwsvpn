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

	socks5 "github.com/armon/go-socks5"
	"github.com/lucasew/wsvpn/pkg/errors"
	"golang.org/x/net/websocket"
)

const (
	wsHttpPort = 3000
)

var (
	err      error
	port     int
	secret   string
	ctx      context.Context
	logfile  *bytes.Buffer
	socksSrv *socks5.Server
)

func init() {
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		errors.ReportError(err, "parsing PORT")
		panic(err)
	}
	secret = os.Getenv("SECRET")
	if secret == "" {
		err := fmt.Errorf("SECRET is not defined")
		errors.ReportError(err, "checking SECRET")
		panic(err)
	}
	logfile = bytes.NewBuffer([]byte{})
	log.SetOutput(logfile)

	socksSrv, err = socks5.New(&socks5.Config{})
	if err != nil {
		errors.ReportError(err, "initializing socks5")
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
		errors.ReportError(err, fmt.Sprintf("SpawnProgram %s", name))
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
	go SpawnProgram("rclone", "serve", "webdav", "--addr", ":9999", "davsrv:/", "--config", "./rclone.conf")
	http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)
	http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(func(conn *websocket.Conn) {
		socksSrv.ServeConn(conn)
	}))
	http.HandleFunc("/", RootHTTPHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		errors.ReportError(err, "http.ListenAndServe")
		panic(err)
	}
}
