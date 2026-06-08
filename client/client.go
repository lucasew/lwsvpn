/*
Package main implements the client-side component of the application.
It listens for local incoming TCP connections on a specified address and proxies them
to a remote SOCKS5 websocket server. This enables local applications to tunnel
traffic securely through the provided remote endpoint.
*/
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

/*
init parses command-line flags to configure the client.
It defines the local listening address ("-addr") and the remote
SOCKS5 websocket server URL ("-srv") to proxy connections towards.
*/
func init() {
    flag.StringVar(&addr, "addr", ":3000", "where to listen for socks5 connections")
    flag.StringVar(&serverURL, "srv", "ws://localhost:1234/test", "where is the websocket server that provides everything")
    flag.Parse()
}

/*
main starts the primary client loop.
It opens a local TCP listener on the configured address and infinitely
accepts incoming connections. Each accepted connection is passed in a new
goroutine to handleConnection along with a newly constructed websocket configuration.
*/
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
        log.Printf("%s connected", conn.RemoteAddr().String())
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

/*
getProxiedConn resolves the proxy connection chain to reach the remote websocket server.
It attempts the following strategies in order:
1. SOCKS5 proxy via environment settings (proxy.FromEnvironment).
2. HTTP CONNECT proxy via environment settings (http.ProxyFromEnvironment).
3. Fallback to a direct TCP dial to the destination host.
*/
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

/*
getWsConfig constructs a standard websocket.Config struct using the provided
serverURL. It requires a mock Origin URL ("http://localhost/") to satisfy
the websocket handshake protocol.
*/
func getWsConfig() (*websocket.Config, error) {
    config, err := websocket.NewConfig(serverURL, "http://localhost/")
    if err != nil {
        return nil, err
    }
    return config, nil
}

/*
handleConnection manages the full lifecycle of a proxied TCP session:
1. Obtains a proxy-routed TCP connection to the remote server via getProxiedConn.
2. Upgrades the TCP connection to a websocket protocol using websocket.NewClient.
3. Launches bidirectional io.Copy operations (via iocopy) to stream data
   between the local client and the remote websocket.
4. Ensures clean channel closure and safely half-closes TCP streams upon completion.
*/
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

/*
closeWrite safely terminates the write-side of a net.Conn.
It type-asserts the connection to the custom `closeable` interface
and invokes CloseWrite() if supported. This is critical for signaling
an EOF condition to the remote peer without tearing down the read channel.
*/
func closeWrite(conn net.Conn) {
	if closeme, ok := conn.(closeable); ok {
		closeme.CloseWrite()
	}
}

/*
iocopy is a wrapper around io.Copy that channels the exit status backwards.
It facilitates synchronizing the termination of the two concurrent streams
used in handleConnection's bidirectional copy operations.
*/
func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
