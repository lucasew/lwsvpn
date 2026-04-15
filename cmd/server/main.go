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
	std_errors "errors"

	"github.com/armon/go-socks5"
	"github.com/lucasew/wsvpn/pkg/errors"
	"golang.org/x/net/websocket"
)

const (
	wsHttpPort        = 3000
	defaultWstunnelURL = "ws://0.0.0.0:3000"
	defaultRcloneAddr = ":9999"
	rcloneWebdavMode  = "webdav"
	rcloneDavsrvRoot  = "davsrv:/"
	rcloneConfigPath  = "./rclone.conf"
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
		errors.ReportError(err)
		panic(err)
	}
	secret = os.Getenv("SECRET")
	if secret == "" {
		err := std_errors.New("SECRET is not defined")
		errors.ReportError(err)
		panic(err)
	}
	logfile = bytes.NewBuffer([]byte{})
	log.SetOutput(logfile)

	socksSrv, err = socks5.New(&socks5.Config{})
	if err != nil {
		errors.ReportError(err)
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
		errors.ReportError(err)
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
	go SpawnProgram("wstunnel", "--server", defaultWstunnelURL)
	go SpawnProgram("rclone", "serve", rcloneWebdavMode, "--addr", defaultRcloneAddr, rcloneDavsrvRoot, "--config", rcloneConfigPath)

	http.HandleFunc(fmt.Sprintf("/%s/log", secret), LogHTTPHandler)
	if err != nil {
		errors.ReportError(err)
		panic(err)
	}

	http.Handle(fmt.Sprintf("/%s", secret), websocket.Handler(func(conn *websocket.Conn) {
		socksSrv.ServeConn(conn)
	}))
	http.HandleFunc("/", RootHTTPHandler)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		errors.ReportError(err)
		panic(err)
	}
}
