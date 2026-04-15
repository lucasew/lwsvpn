package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/lucasew/wsvpn/pkg/errors"
	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
)

const (
	defaultListenAddr = ":3000"
	defaultServerURL  = "ws://localhost:1234/test"
	defaultOrigin     = "http://localhost/"
)

var (
	addr      string
	serverURL string
)

func init() {
	flag.StringVar(&addr, "addr", defaultListenAddr, "where to listen for socks5 connections")
	flag.StringVar(&serverURL, "srv", defaultServerURL, "where is the websocket server that provides everything")
	flag.Parse()
}

func main() {
	log.Printf("initializing...")
	log.Printf("listening socks5 @ %s...", addr)
	log.Printf("using server %s...", serverURL)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		errors.ReportError(fmt.Errorf("error listening: %w", err))
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			errors.ReportError(fmt.Errorf("error accepting connection: %w", err))
			continue
		}
		log.Printf("%s connected", conn.RemoteAddr().String())
		cfg, err := getWsConfig()
		if err != nil {
			errors.ReportError(fmt.Errorf("error ws config: %w", err))
			conn.Close()
			continue
		}
		go handleConnection(cfg, conn)
	}
}

func getProxiedConn(turl url.URL) (net.Conn, error) {
	// We first try to get a Socks5 proxied connection. If that fails, we're moving on to http{s,}_proxy.
	dialer := proxy.FromEnvironment()
	if dialer != proxy.Direct {
		return dialer.Dial("tcp", turl.Host)
	}

	turl.Scheme = strings.Replace(turl.Scheme, "ws", "http", 1)
	proxyURL, err := http.ProxyFromEnvironment(&http.Request{URL: &turl})
	if proxyURL == nil {
		return net.Dial("tcp", turl.Host)
	}

	p, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		return nil, err
	}

	cc := httputil.NewProxyClientConn(p, nil)
	_, err = cc.Do(&http.Request{
		Method: "CONNECT",
		URL:    &url.URL{},
		Host:   turl.Host,
	})
	if err != nil && err != httputil.ErrPersistEOF {
		return nil, err
	}

	conn, _ := cc.Hijack()

	return conn, nil
}

func getWsConfig() (*websocket.Config, error) {
	config, err := websocket.NewConfig(serverURL, defaultOrigin)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func handleConnection(wsConfig *websocket.Config, conn net.Conn) {
	defer conn.Close()

	tcp, err := getProxiedConn(*wsConfig.Location)
	if err != nil {
		errors.ReportError(fmt.Errorf("getProxiedConn(): %w", err))
		return
	}

	ws, err := websocket.NewClient(wsConfig, tcp)
	if err != nil {
		errors.ReportError(fmt.Errorf("websocket.NewClient(): %w", err))
		return
	}
	defer ws.Close()

	c := make(chan error, 2)
	go iocopy(ws, conn, c)
	go iocopy(conn, ws, c)

	for i := 0; i < 2; i++ {
		if err := <-c; err != nil {
			errors.ReportError(fmt.Errorf("io.Copy(): %w", err))
			return
		}
		// If any of the sides closes the connection, we want to close the write channel.
		closeWrite(conn)
		closeWrite(tcp)
	}
}

type closeable interface {
	CloseWrite() error
}

func closeWrite(conn net.Conn) {
	if closeme, ok := conn.(closeable); ok {
		closeme.CloseWrite()
	}
}

func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
