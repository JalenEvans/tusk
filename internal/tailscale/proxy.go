package tailscale

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Proxy is a SOCKS5 proxy server that listens on a specified address.
type Proxy struct {
	addr     string
	mu       sync.Mutex
	listener net.Listener
}

func NewProxy(addr string) *Proxy {
	return &Proxy{addr: addr}
}

// Start begins listening for connections on the configured address.
// It returns immediately after the listener is established; connections
// are accepted in a background goroutine.
func (p *Proxy) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("proxy listen failed: %w", err)
	}

	p.mu.Lock()
	p.listener = ln
	p.mu.Unlock()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleSOCKSConnection(conn)
		}
	}()

	return nil
}

// Stop closes the listener and releases the port.
// It is safe to call Stop multiple times.
func (p *Proxy) Stop() error {
	p.mu.Lock()
	ln := p.listener
	p.listener = nil
	p.mu.Unlock()

	if ln != nil {
		return ln.Close()
	}
	return nil
}

// handleSOCKSConnection handles a single SOCKS5 client connection.
// For now it performs a minimal handshake and closes the connection.
func handleSOCKSConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	conn.Read(buf)
}
