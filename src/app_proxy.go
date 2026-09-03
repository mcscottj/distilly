package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"distilly/internal/proxy"
)

const (
	settingProxyPort = "proxy_port"
	defaultProxyPort = "8787"
)

// StartProxy starts the local OpenAI-compatible proxy on 127.0.0.1 using the
// configured proxy_port setting (default 8787).
func (a *App) StartProxy() error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}

	port, err := a.resolveProxyPort()
	if err != nil {
		return err
	}
	addr := proxy.DefaultListenHost + ":" + port

	a.proxyMu.Lock()
	defer a.proxyMu.Unlock()

	if a.proxy != nil && a.proxy.Running() {
		return fmt.Errorf("proxy already running on %s", a.proxy.Addr())
	}

	p := proxy.New(s)
	if err := p.Start(addr); err != nil {
		return err
	}
	a.proxy = p
	return nil
}

// StopProxy stops the local proxy if it is running. Idempotent.
func (a *App) StopProxy() error {
	a.proxyMu.Lock()
	defer a.proxyMu.Unlock()

	if a.proxy == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := a.proxy.Shutdown(ctx)
	a.proxy = nil
	return err
}

// GetProxyStatus returns whether the proxy is listening and its address.
func (a *App) GetProxyStatus() proxy.Status {
	a.proxyMu.Lock()
	defer a.proxyMu.Unlock()
	if a.proxy == nil {
		return proxy.Status{}
	}
	return a.proxy.Status()
}

func (a *App) resolveProxyPort() (string, error) {
	s, err := a.requireStore()
	if err != nil {
		return "", err
	}
	port, err := s.GetSetting(settingProxyPort)
	if err != nil {
		return "", err
	}
	port = strings.TrimSpace(port)
	if port == "" {
		port = defaultProxyPort
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid proxy port %q", port)
	}
	return strconv.Itoa(n), nil
}
