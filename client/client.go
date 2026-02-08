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
    "sync"
    "time"

	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
    "github.com/lucasew/wsvpn/pkg/errors"
    "github.com/hashicorp/yamux"
)

var (
    addr string
    serverURL string
    session *yamux.Session
    sessionLock sync.Mutex
)

func init() {
    flag.StringVar(&addr, "addr", ":3000", "where to listen for socks5 connections")
    flag.StringVar(&serverURL, "srv", "ws://localhost:1234/test", "where is the websocket server that provides everything")
    flag.Parse()
}

func main() {
    defer errors.ReportPanic()
    log.Printf("initializing...")
    log.Printf("listening socks5 @ %s...", addr)
    log.Printf("using server %s...", serverURL)

    go manageSession()

    l, err := net.Listen("tcp", addr)
    if err != nil {
        errors.ReportError(err, "Failed to listen")
        panic(err)
    }
    for {
        conn, err := l.Accept()
        log.Printf("%s connected", conn.RemoteAddr().String())
        if err != nil {
            log.Printf("error accepting connection: %s", err.Error())
            errors.ReportError(err, "Error accepting connection")
            continue
        }
        go handleConnection(conn)
    }
}

func manageSession() {
    for {
        // log.Println("Connecting to server...") // Too verbose if looping fast?
        cfg, err := getWsConfig()
        if err != nil {
            log.Printf("error ws config: %s", err.Error())
            errors.ReportError(err, "ws config error")
            time.Sleep(5 * time.Second)
            continue
        }

        tcp, err := getProxiedConn(*cfg.Location)
        if err != nil {
            log.Printf("getProxiedConn(): %s", err)
            errors.ReportError(err, "getProxiedConn failed")
            time.Sleep(5 * time.Second)
            continue
        }

        ws, err := websocket.NewClient(cfg, tcp)
        if err != nil {
            log.Printf("websocket.NewClient(): %s", err)
            errors.ReportError(err, "websocket.NewClient failed")
            tcp.Close()
            time.Sleep(5 * time.Second)
            continue
        }

        // Use default config with keepalive
        conf := yamux.DefaultConfig()
        conf.KeepAliveInterval = 30 * time.Second

        sess, err := yamux.Client(ws, conf)
        if err != nil {
            errors.ReportError(err, "yamux client creation failed")
            ws.Close()
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Session established")
        sessionLock.Lock()
        session = sess
        sessionLock.Unlock()

        // Wait until session is closed
        for !sess.IsClosed() {
            time.Sleep(1 * time.Second)
        }

        log.Println("Session disconnected")
        sessionLock.Lock()
        session = nil
        sessionLock.Unlock()
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

func handleConnection(conn net.Conn) {
	defer conn.Close()

    sessionLock.Lock()
    sess := session
    sessionLock.Unlock()

    if sess == nil || sess.IsClosed() {
        log.Println("No active session")
        // Optionally wait for session?
        // For now, fail fast.
        return
    }

	stream, err := sess.Open()
	if err != nil {
		log.Print("yamux.Session.Open(): ", err)
        errors.ReportError(err, "yamux open stream failed")
		return
	}
	defer stream.Close()

	c := make(chan error, 2)
	go iocopy(stream, conn, c)
	go iocopy(conn, stream, c)

	for i := 0; i < 2; i++ {
		if err := <-c; err != nil {
			log.Printf("io.Copy(): %s", err.Error())
            // errors.ReportError(err, "io.Copy failed")
			return
		}
		// If any of the sides closes the connection, we want to close the write channel.
		closeWrite(conn)
		closeWrite(stream)
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
