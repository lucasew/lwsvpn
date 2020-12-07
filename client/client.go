package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
)

var (
    addr string
    serverURL string
)

func init() {
    flag.StringVar(&addr, "addr", ":3000", "where to listen for socks5 connections")
    flag.StringVar(&serverURL, "srv", "ws://localhost:1234/test", "where is the websocket server that provides everything")
    flag.Parse()
}

func main() {
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
        cfg, err := getWsConfig()
        if err != nil {
            log.Printf("error ws config: %s", err.Error())
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

func handleConnection(wsConfig *websocket.Config, conn net.Conn) {
	defer conn.Close()

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
	defer ws.Close()

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
		closeme.CloseWrite()
	}
}

func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
