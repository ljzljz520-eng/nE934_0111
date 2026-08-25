package httpapi

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/flow026"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
	"tastinginvite/internal/workflow"
	"testing"
	"time"
)

func TestHTTPHealthAndCreate(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := clock.Fixed{At: time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)}
	r := repository.New(s, c)
	server := New(r, workflow.New(r, c), flow026.New(r, c))
	health := httptest.NewRecorder()
	server.Mux.ServeHTTP(health, httptest.NewRequest("GET", "/health", nil))
	if health.Code != 200 || !strings.Contains(health.Body.String(), "ok") {
		t.Fatalf("health=%d %s", health.Code, health.Body.String())
	}
	body := strings.NewReader(`{"id":"http-1","title":"Web Tasting","host":"Mina","venue":"Hall","startAt":"2030-01-01T10:00:00Z","endAt":"2030-01-01T12:00:00Z","capacity":4}`)
	created := httptest.NewRecorder()
	server.Mux.ServeHTTP(created, httptest.NewRequest("POST", "/invitations", body))
	if created.Code != 201 {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
}
