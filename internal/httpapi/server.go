package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"tastinginvite/internal/flow026"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/workflow"
)

type Server struct {
	Repo     *repository.Repository
	Flow     *workflow.Engine
	Exporter *flow026.Handler
	Mux      *http.ServeMux
}

func New(repo *repository.Repository, flow *workflow.Engine, exporter *flow026.Handler) *Server {
	s := &Server{Repo: repo, Flow: flow, Exporter: exporter, Mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Mux.HandleFunc("/health", s.health)
	s.Mux.HandleFunc("/invitations", s.invitations)
	s.Mux.HandleFunc("/invitations/", s.invitation)
	s.Mux.HandleFunc("/exports", s.exports)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) invitations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		records, err := s.Repo.Search(model.InvitationFilter{Query: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status"), Host: r.URL.Query().Get("host")})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, records)
	case http.MethodPost:
		var record model.Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			writeErrorStatus(w, http.StatusBadRequest, err)
			return
		}
		created, err := s.Repo.Create(record, actor(r))
		if err != nil {
			writeErrorStatus(w, http.StatusUnprocessableEntity, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) invitation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/invitations/")
	if id == "" {
		writeErrorStatus(w, http.StatusBadRequest, errors.New("id required"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		record, err := s.Repo.Get(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, record)
	case http.MethodPatch:
		var input struct {
			Status string `json:"status"`
			Actor  string `json:"actor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeErrorStatus(w, http.StatusBadRequest, err)
			return
		}
		record, err := s.Repo.Transition(id, input.Status, input.Actor)
		if err != nil {
			writeErrorStatus(w, http.StatusConflict, err)
			return
		}
		writeJSON(w, http.StatusOK, record)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) exports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	result, err := s.Exporter.Export(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func actor(r *http.Request) string {
	value := r.Header.Get("X-Actor")
	if value == "" {
		return "system"
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) { writeErrorStatus(w, http.StatusNotFound, err) }

func writeErrorStatus(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
