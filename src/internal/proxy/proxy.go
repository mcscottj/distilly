// Package proxy provides an OpenAI-compatible local chat completions
// gateway that lint-optimizes prompts before forwarding upstream.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"distilly/internal/lint"
	"distilly/internal/store"
	"distilly/internal/tokenizer"
)

// Setting keys (must match frontend SettingKey / store values).
const (
	SettingUpstreamURL           = "upstream_url"
	SettingAPIKey                = "api_key"
	SettingApproveNearDuplicates = "approve_near_duplicates"
	SettingApproveJSONConversion = "approve_json_conversion"
	SettingPassthrough           = "passthrough"
	DefaultUpstreamURL           = "https://api.openai.com"
)

// DefaultListenHost is the loopback host used when starting from the desktop app.
const DefaultListenHost = "127.0.0.1"

// Status reports whether the local proxy is listening and on which address.
type Status struct {
	Running bool   `json:"running"`
	Addr    string `json:"addr"`
}

// Server is an OpenAI-compatible proxy that optimizes chat prompts.
type Server struct {
	store  *store.Store
	client *http.Client

	mu      sync.Mutex
	server  *http.Server
	addr    string
	running bool
}

// New returns a proxy Server that reads settings and logs via s.
func New(s *store.Store) *Server {
	return &Server{
		store: s,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Handler returns the HTTP handler for POST /v1/chat/completions.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	return mux
}

// Start binds addr and serves in the background. Returns once listening
// succeeds (or fails to bind). Calling Start while already running is an error.
func (s *Server) Start(addr string) error {
	s.mu.Lock()
	if s.running {
		cur := s.addr
		s.mu.Unlock()
		return fmt.Errorf("proxy already running on %s", cur)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	// Prefer the concrete bound address (handles ":0" in tests).
	bound := ln.Addr().String()
	s.server = srv
	s.addr = bound
	s.running = true
	s.mu.Unlock()

	go func() {
		err := srv.Serve(ln)
		s.mu.Lock()
		s.running = false
		s.server = nil
		s.addr = ""
		s.mu.Unlock()
		_ = err // ErrServerClosed is expected on Shutdown
	}()
	return nil
}

// ListenAndServe starts the proxy on addr (e.g. "127.0.0.1:8787") and blocks.
func (s *Server) ListenAndServe(addr string) error {
	if err := s.Start(addr); err != nil {
		return err
	}
	// Block until Shutdown clears running state.
	for {
		s.mu.Lock()
		running := s.running
		s.mu.Unlock()
		if !running {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Status returns the current listen state.
func (s *Server) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Status{Running: s.running, Addr: s.addr}
}

// Running reports whether the server is currently serving.
func (s *Server) Running() bool {
	return s.Status().Running
}

// Addr returns the bound listen address when running, otherwise "".
func (s *Server) Addr() string {
	return s.Status().Addr
}

// Shutdown gracefully stops the HTTP server if running.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	err := srv.Shutdown(ctx)
	s.mu.Lock()
	s.running = false
	s.server = nil
	s.addr = ""
	s.mu.Unlock()
	return err
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type runtimeConfig struct {
	UpstreamURL string
	APIKey      string
	ApplyOpts   lint.ApplyOptions
	Passthrough bool

	RepoRoot           string
	CodeContextEnabled bool
	ContextMaxDepth    int
	ContextMaxTokens   int
}

func (s *Server) loadConfig() (runtimeConfig, error) {
	cfg := runtimeConfig{UpstreamURL: DefaultUpstreamURL}
	if s.store == nil {
		return cfg, fmt.Errorf("store not available")
	}

	upstream, err := s.store.GetSetting(SettingUpstreamURL)
	if err != nil {
		return cfg, err
	}
	if strings.TrimSpace(upstream) != "" {
		cfg.UpstreamURL = strings.TrimRight(strings.TrimSpace(upstream), "/")
	}

	cfg.APIKey, err = s.store.GetSetting(SettingAPIKey)
	if err != nil {
		return cfg, err
	}

	near, err := s.store.GetSetting(SettingApproveNearDuplicates)
	if err != nil {
		return cfg, err
	}
	jsonConv, err := s.store.GetSetting(SettingApproveJSONConversion)
	if err != nil {
		return cfg, err
	}
	pass, err := s.store.GetSetting(SettingPassthrough)
	if err != nil {
		return cfg, err
	}

	cfg.ApplyOpts = lint.ApplyOptions{
		ApproveNearDuplicates: parseBool(near),
		ApproveJSONConversion: parseBool(jsonConv),
	}
	cfg.Passthrough = parseBool(pass)

	repoRoot, err := s.store.GetSetting(SettingRepoRoot)
	if err != nil {
		return cfg, err
	}
	cfg.RepoRoot = strings.TrimSpace(repoRoot)

	enabled, err := s.store.GetSetting(SettingCodeContextEnabled)
	if err != nil {
		return cfg, err
	}
	cfg.CodeContextEnabled = parseBool(enabled)

	depth, err := s.store.GetSetting(SettingContextMaxDepth)
	if err != nil {
		return cfg, err
	}
	cfg.ContextMaxDepth = parseContextIntSetting(depth, DefaultContextMaxDepth)

	maxTok, err := s.store.GetSetting(SettingContextMaxTokens)
	if err != nil {
		return cfg, err
	}
	cfg.ContextMaxTokens = parseContextIntSetting(maxTok, DefaultContextMaxTokens)

	return cfg, nil
}

func parseBool(v string) bool {
	return v == "true" || v == "1"
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var req chatRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Stream {
		_ = s.logRejectedStream(req)
		http.Error(w, "streaming not supported yet; set stream=false", http.StatusBadRequest)
		return
	}

	cfg, err := s.loadConfig()
	if err != nil {
		http.Error(w, "proxy misconfigured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prompt := MessagesToPrompt(req.Messages)
	prompt = applyCodeContext(prompt, cfg)
	report := lint.Run(prompt, req.Model)
	optimizedPrompt := lint.Apply(prompt, cfg.ApplyOpts)
	optimizedTokens := tokenizer.Count(optimizedPrompt)

	forwardBody := raw
	if !cfg.Passthrough {
		out := req
		out.Messages = PromptToMessages(optimizedPrompt)
		out.Stream = false
		encoded, err := json.Marshal(out)
		if err != nil {
			http.Error(w, "failed to encode optimized request", http.StatusInternalServerError)
			return
		}
		forwardBody = encoded
	}

	_ = s.logRequest(req.Model, report, optimizedTokens)

	upstreamURL := cfg.UpstreamURL + "/v1/chat/completions"
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(forwardBody))
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		upReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	// Forward optional OpenAI org/project headers when present.
	for _, h := range []string{"OpenAI-Organization", "OpenAI-Project"} {
		if v := r.Header.Get(h); v != "" {
			upReq.Header.Set(h, v)
		}
	}

	resp, err := s.client.Do(upReq)
	if err != nil {
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) logRequest(model string, report lint.Report, optimizedTokens int) error {
	if s.store == nil {
		return nil
	}
	savingsPct := 0.0
	if report.InputTokens > 0 {
		saved := report.InputTokens - optimizedTokens
		if saved < 0 {
			saved = 0
		}
		savingsPct = float64(saved) / float64(report.InputTokens) * 100
	}
	_, err := s.store.InsertRequest(store.Request{
		Source:          store.SourceProxy,
		Model:           model,
		InputTokens:     report.InputTokens,
		OptimizedTokens: optimizedTokens,
		SavingsPct:      savingsPct,
		CostUSD:         report.EstimatedCostUSD,
		SavingsUSD:      report.EstimatedSavingsUSD,
	})
	return err
}

func (s *Server) logRejectedStream(req chatRequest) error {
	if s.store == nil {
		return nil
	}
	report := lint.Run(MessagesToPrompt(req.Messages), req.Model)
	_, err := s.store.InsertRequest(store.Request{
		Source:          store.SourceProxyStream,
		Model:           req.Model,
		InputTokens:     report.InputTokens,
		OptimizedTokens: report.InputTokens,
		SavingsPct:      0,
		CostUSD:         report.EstimatedCostUSD,
		SavingsUSD:      0,
	})
	return err
}
