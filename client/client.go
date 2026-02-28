package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"

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
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("error accepting connection: %s", err.Error())
			continue
		}
		log.Printf("%s connected", conn.RemoteAddr().String())

		cfg, err := getWsConfig()
		if err != nil {
			log.Printf("error ws config: %s", err.Error())
			errors.ReportError(conn.Close())
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

	req := &http.Request{URL: &turl}
	proxyURL, err := http.ProxyFromEnvironment(req)
	if err != nil || proxyURL == nil {
		return net.Dial("tcp", turl.Host)
	}

	p, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		return nil, err
	}

	req = &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: turl.Host},
		Host:   turl.Host,
	}

	err = req.Write(p)
	if err != nil {
		errors.ReportError(p.Close())
		return nil, err
	}

	return p, nil
}

func getWsConfig() (*websocket.Config, error) {
	config, err := websocket.NewConfig(serverURL, "http://localhost/")
	if err != nil {
		return nil, err
	}
	return config, nil
}

func handleConnection(wsConfig *websocket.Config, conn net.Conn) {
	defer func() {
		errors.ReportError(conn.Close())
	}()

	tcp, err := getProxiedConn(*wsConfig.Location)
	if err != nil {
		log.Print("getProxiedConn(): ", err)
		return
	}

	ws, err := websocket.NewClient(wsConfig, tcp)
	if err != nil {
		log.Print("websocket.NewClient(): ", err)
		return
	}
	defer func() {
		errors.ReportError(ws.Close())
	}()

	c := make(chan error, 2)
	go iocopy(ws, conn, c)
	go iocopy(conn, ws, c)

	for i := 0; i < 2; i++ {
		if err := <-c; err != nil {
			log.Printf("io.Copy(): %s", err.Error())
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
		errors.ReportError(closeme.CloseWrite())
	}
}

func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
