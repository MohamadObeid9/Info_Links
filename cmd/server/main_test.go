package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestSmokeHealthzAndGracefulShutdown(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	server := newHTTPServer(addr, mux)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- serveHTTP(ctx, server, ln)
	}()

	waitForServer(t, "http://"+addr+"/healthz")

	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		cancel()
		t.Fatalf("GET /healthz: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("GET /healthz status: got %d want 200", resp.StatusCode)
	}
	if string(body) != "ok" {
		cancel()
		t.Fatalf("GET /healthz body: got %q want ok", body)
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serveHTTP: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s")
	}
}

func TestNewHTTPServerTimeouts(t *testing.T) {
	t.Parallel()

	srv := newHTTPServer(":0", http.NewServeMux())
	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout: got %v want 5s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 15*time.Second {
		t.Fatalf("ReadTimeout: got %v want 15s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 60*time.Second {
		t.Fatalf("WriteTimeout: got %v want 60s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout: got %v want 60s", srv.IdleTimeout)
	}
}

func waitForServer(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server not ready at %s", url)
}
