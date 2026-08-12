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

// Server is an OpenAI-compatible proxy that optimizes chat prompts.
type Server struct {
	store  *store.Store
	client *http.Client

	mu     sync.Mutex
	server *http.Server
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

// ListenAndServe starts the proxy on addr (e.g. "127.0.0.1:8787").
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.mu.Lock()
	s.server = srv
	s.mu.Unlock()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

// Shutdown gracefully stops the HTTP server if running.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
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
		http.Error(w, "streaming not supported yet; set stream=false", http.StatusBadRequest)
		return
	}

	cfg, err := s.loadConfig()
	if err != nil {
		http.Error(w, "proxy misconfigured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prompt := MessagesToPrompt(req.Messages)
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
