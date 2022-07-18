// Package optls lets you accept TLS and unencrypted connections on the same port.
package optls

import (
	"crypto/tls"
	"io"
	"net"
)

// Listen creates a listener accepting connections on the given network address using [net.Listen].
func Listen(network, address string, config *tls.Config) (net.Listener, error) {
	inner, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(inner, config), nil
}

// NewListener creates a Listener which accepts connections from an inner Listener.
// If config is valid, and the client sends a ClientHello message,
// the connection is wrapped with a [tls.Server].
func NewListener(inner net.Listener, config *tls.Config) net.Listener {
	if config == nil || len(config.Certificates) == 0 &&
		config.GetCertificate == nil && config.GetConfigForClient == nil {
		return inner
	}
	return &listener{
		Listener: inner,
		config:   config,
	}
}

type listener struct {
	net.Listener
	config *tls.Config
}

func (l *listener) Accept() (net.Conn, error) {
	inner, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	conn := conn{Conn: inner}
	conn.n, conn.err = inner.Read(conn.p[:])
	if conn.n == 1 && conn.p[0] == 0x16 {
		return tls.Server(&conn, l.config), nil
	}
	return &conn, nil
}

type conn struct {
	net.Conn
	p    [1]byte
	n    int
	err  error
	done bool
}

func (c *conn) Read(b []byte) (int, error) {
	if !c.done && len(b) > 0 {
		c.done = true
		b[0] = c.p[0]
		return c.n, c.err
	}
	return c.Conn.Read(b)
}

func (c *conn) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := c.Conn.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(c.Conn, r)
}

func (c *conn) Close() error {
	c.done = true
	return c.Conn.Close()
}
