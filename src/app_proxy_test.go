package main

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	"distilly/internal/store"
)

func TestAppProxyLifecycle(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "proxy-app.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app.store = s

	// Bind an ephemeral port so tests do not collide.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	_ = ln.Close()

	if err := app.SetSetting("proxy_port", port); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}

	st := app.GetProxyStatus()
	if st.Running {
		t.Fatalf("status before start = %+v", st)
	}

	if err := app.StartProxy(); err != nil {
		t.Fatalf("StartProxy: %v", err)
	}
	t.Cleanup(func() { _ = app.StopProxy() })

	st = app.GetProxyStatus()
	if !st.Running {
		t.Fatalf("status after start = %+v", st)
	}
	if st.Addr == "" {
		t.Fatal("expected non-empty addr")
	}

	if err := app.StartProxy(); err == nil {
		t.Fatal("expected error starting proxy twice")
	}

	if err := app.StopProxy(); err != nil {
		t.Fatalf("StopProxy: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for app.GetProxyStatus().Running && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	st = app.GetProxyStatus()
	if st.Running {
		t.Fatalf("status after stop = %+v", st)
	}

	// Idempotent stop.
	if err := app.StopProxy(); err != nil {
		t.Fatalf("StopProxy again: %v", err)
	}
}

func TestAppStartProxyInvalidPort(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "proxy-port.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app.store = s

	if err := app.SetSetting("proxy_port", "not-a-port"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := app.StartProxy(); err == nil {
		t.Fatal("expected invalid port error")
	}
}
