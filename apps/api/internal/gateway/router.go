package gateway

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleLLMGateway(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/v1/chat/completions":
		s.handleOpenAIChat(w, r)
	case path == "/v1/messages":
		s.handleAnthropicMessages(w, r)
	default:
		writeError(w, http.StatusNotFound, "not_found", "Unknown endpoint")
	}
}

func (s *Server) handleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	s.HandleOpenAIChat(w, r)
}

func (s *Server) handleAnthropicMessages(w http.ResponseWriter, r *http.Request) {
	s.HandleAnthropicMessages(w, r)
}

type Server struct {
	httpServer *http.Server
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	srv := &Server{}

	mux.HandleFunc("/v1/chat/completions", srv.handleLLMGateway)
	mux.HandleFunc("/v1/messages", srv.handleLLMGateway)

	srv.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * 0,
	}

	return srv
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
