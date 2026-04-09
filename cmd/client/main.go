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
	"os"
	"strings"

	"github.com/lucasew/wsvpn/pkg/errors"

	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
)

var (
	addr      string
	serverURL string
)

func init() {
	flag.StringVar(&addr, "addr", ":3000", "where to listen for socks5 connections")
	flag.StringVar(&serverURL, "srv", "ws://localhost:1234/test", "where is the websocket server that provides everything")
	flag.Parse()
}

func main() {
	log.Printf("initializing...")
	log.Printf("listening socks5 @ %s...", addr)
	log.Printf("using server %s...", serverURL)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		errors.ReportError(fmt.Errorf("failed to listen on tcp %s: %w", addr, err))
		os.Exit(1)
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
			errors.ReportError(fmt.Errorf("error building ws config: %w", err))
			conn.Close()
			continue
		}

		go handleConnection(cfg, conn)
	}
}

func getProxiedConn(turl url.URL) (net.Conn, error) {
	// We first try to get a Socks5 proxied conncetion. If that fails, we're moving on to http{s,}_proxy.
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
	config, err := websocket.NewConfig(serverURL, "http://localhost/")
	if err != nil {
		return nil, err
	}
	return config, nil
}

const numCopyChannels = 2

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

	c := make(chan error, numCopyChannels)
	go iocopy(ws, conn, c)
	go iocopy(conn, ws, c)

	for i := 0; i < numCopyChannels; i++ {
		if err := <-c; err != nil {
			errors.ReportError(fmt.Errorf("io.Copy() during stream replication: %w", err))
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
		err := closeme.CloseWrite()
		if err != nil {
			errors.ReportError(fmt.Errorf("failed to CloseWrite connection: %w", err))
		}
	}
}

func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
