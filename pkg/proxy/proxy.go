package proxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bobcob7/sni-proxy/internal/proxy"
	"go.uber.org/zap"
)

type Proxy struct {
	Address string
	Dialer  struct {
		Timeout   time.Duration
		Deadline  time.Duration
		Keepalive time.Duration
	}
	dialer   net.Dialer
	Upstream map[string]string
}

func (p *Proxy) Run(ctx context.Context, logger *zap.Logger) {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", p.Address)
	if err != nil {
		logger.Error("failed to listen for connection", zap.String("address", p.Address), zap.Error(err))
	}
	pendingConnections := make(chan net.Conn, 10)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warn("failed to accept connection", zap.Error(err))
				return
			}
			pendingConnections <- conn
		}
	}()
	logger.Info("listing for connection", zap.String("address", p.Address))
handlerLoop:
	for {
		select {
		case <-ctx.Done():
			break handlerLoop
		case conn := <-pendingConnections:
			handleLogger := logger.With(zap.String("localAddr", conn.LocalAddr().String()), zap.String("remoteAddr", conn.RemoteAddr().String()))
			go p.handle(ctx, handleLogger, conn)
		}
	}
}

func (u *Proxy) handle(ctx context.Context, logger *zap.Logger, incomingConn net.Conn) {
	defer incomingConn.Close()
	// Try to get SNI
	helloBuffer := make([]byte, 4096)
	helloBufferSize, err := incomingConn.Read(helloBuffer)
	if err != nil {
		logger.Error("failed to read incoming connection", zap.Error(err))
	}
	sni, err := proxy.GetSNI(helloBuffer)
	if err != nil {
		logger.Warn("could not find SNI for connection")
		return
	}
	logger = logger.With(zap.String("sni", sni))
	logger.Debug("found SNI")
	// Get connection to upstream
	upstreamAddress := u.Upstream[sni]
	upstreamConn, err := u.dialer.DialContext(ctx, "tcp", upstreamAddress)
	if err != nil {
		logger.Warn("could not find upstream for connection")
		return
	}
	defer upstreamConn.Close()
	// Send handshake message upstream
	logger.Info("proxying to server", zap.String("upstream", upstreamAddress))
	if _, err := upstreamConn.Write(helloBuffer[:helloBufferSize]); err != nil {
		logger.Error("could not write extra handshake upstream")
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if n, err := io.Copy(upstreamConn, incomingConn); err != nil {
			logger.Error("failed to copy upstream -> incoming", zap.Error(err))
		} else {
			logger.Debug("copied upstream -> incoming", zap.Int64("bytes", n))
		}
	}()
	go func() {
		defer wg.Done()
		if n, err := io.Copy(incomingConn, upstreamConn); err != nil {
			logger.Error("failed to copy incoming -> upstream", zap.Error(err))
		} else {
			logger.Debug("copied incoming -> upstream", zap.Int64("bytes", n))
		}
	}()
	wg.Wait()
	logger.Debug("finished handling connection")
}
