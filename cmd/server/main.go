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

	"github.com/lucasew/wsvpn/pkg/errors"

	socks5 "github.com/armon/go-socks5"
	"golang.org/x/net/websocket"
)

const (
	wsTunnelPort = 3000
	rclonePort   = 9999
)

type Server struct {
	port     int
	secret   string
	logfile  *bytes.Buffer
	socksSrv *socks5.Server
}

func NewServer() (*Server, error) {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT environment variable %q: %w", portStr, err)
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SECRET environment variable is not defined")
	}

	logfile := bytes.NewBuffer([]byte{})
	log.SetOutput(logfile)

	socksSrv, err := socks5.New(&socks5.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 server: %w", err)
	}

	return &Server{
		port:     port,
		secret:   secret,
		logfile:  logfile,
		socksSrv: socksSrv,
	}, nil
}

func (s *Server) SpawnProgram(name string, args ...string) {
	fmt.Fprintf(s.logfile, "spawning: %s %+v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = s.logfile
	cmd.Stderr = s.logfile
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("program %s exited with error: %w", name, err)
		errors.ReportError(err)
		fmt.Fprintf(s.logfile, "%s\n", err.Error())
	}
}

func (s *Server) RootHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello, world")
}

func (s *Server) LogHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, s.logfile)
	if err != nil {
		errors.ReportError(fmt.Errorf("failed to copy logfile: %w", err))
	}
}

func main() {
	srv, err := NewServer()
	if err != nil {
		errors.ReportError(err)
		os.Exit(1)
	}

	go srv.SpawnProgram("wstunnel", "--server", fmt.Sprintf("ws://0.0.0.0:%d", wsTunnelPort))
	go srv.SpawnProgram("rclone", "serve", "webdav", "--addr", fmt.Sprintf(":%d", rclonePort), "davsrv:/", "--config", "./rclone.conf")

	http.HandleFunc(fmt.Sprintf("/%s/log", srv.secret), srv.LogHTTPHandler)
	http.Handle(fmt.Sprintf("/%s", srv.secret), websocket.Handler(func(conn *websocket.Conn) {
		err := srv.socksSrv.ServeConn(conn)
		if err != nil {
			errors.ReportError(fmt.Errorf("socks5 ServeConn error: %w", err))
		}
	}))
	http.HandleFunc("/", srv.RootHTTPHandler)

	err = http.ListenAndServe(fmt.Sprintf(":%d", srv.port), nil)
	if err != nil {
		errors.ReportError(fmt.Errorf("server ListenAndServe error: %w", err))
		os.Exit(1)
	}
}
