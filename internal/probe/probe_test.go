package probe_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/probe"
)

func startTCPListener(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return ln.Addr().String()
}

func TestProbe_TCP_Healthy(t *testing.T) {
	addr := startTCPListener(t)
	p := probe.New(probe.KindTCP, addr, time.Second)
	r := p.Run(context.Background())
	if !r.Healthy {
		t.Fatalf("expected healthy, got err: %v", r.Err)
	}
	if r.Latency <= 0 {
		t.Fatal("expected positive latency")
	}
}

func TestProbe_TCP_Unhealthy(t *testing.T) {
	// Nothing listening on this port.
	p := probe.New(probe.KindTCP, "127.0.0.1:1", 200*time.Millisecond)
	r := p.Run(context.Background())
	if r.Healthy {
		t.Fatal("expected unhealthy")
	}
	if r.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestProbe_HTTP_Healthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	p := probe.New(probe.KindHTTP, ts.URL, time.Second)
	r := p.Run(context.Background())
	if !r.Healthy {
		t.Fatalf("expected healthy, got err: %v", r.Err)
	}
}

func TestProbe_HTTP_Non2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(ts.Close)

	p := probe.New(probe.KindHTTP, ts.URL, time.Second)
	r := p.Run(context.Background())
	if r.Healthy {
		t.Fatal("expected unhealthy for 503")
	}
}

func TestProbe_UnknownKind(t *testing.T) {
	p := probe.New("grpc", "localhost:9090", time.Second)
	r := p.Run(context.Background())
	if r.Healthy {
		t.Fatal("expected unhealthy for unknown kind")
	}
}

func TestProbe_DefaultTimeout(t *testing.T) {
	// Zero timeout should be replaced by the 5 s default without panicking.
	p := probe.New(probe.KindTCP, "127.0.0.1:1", 0)
	r := p.Run(context.Background())
	if r.Healthy {
		t.Fatal("expected unhealthy")
	}
}
