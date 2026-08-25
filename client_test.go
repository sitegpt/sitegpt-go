package sitegpt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient("test-token", WithBaseURL(server.URL))
}

func TestRequestSendsAuthAndDecodes(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.URL.Path; got != "/api/v2/chatbots" {
			t.Errorf("path = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{"chatbots": []any{}}})
	})
	result, err := client.Chatbots.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("ok = %v", result["ok"])
	}
}

func TestErrorsCarryCodeAndHint(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": map[string]any{"code": "CHATBOT_NOT_FOUND", "message": "No such chatbot", "hint": "List chatbots first."},
		})
	})
	_, err := client.Chatbots.Get(context.Background(), "missing")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.Code != "CHATBOT_NOT_FOUND" || apiErr.Status != 404 || apiErr.Hint == "" {
		t.Errorf("unexpected error: %+v", apiErr)
	}
}

func TestDestructiveCallsRequireConfirmation(t *testing.T) {
	called := false
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := r.URL.Query().Get("confirm"); got != "true" {
			t.Errorf("confirm query = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	_, err := client.Chatbots.Delete(context.Background(), "123", false)
	apiErr, ok := err.(*Error)
	if !ok || apiErr.Code != "CONFIRMATION_REQUIRED" {
		t.Fatalf("expected CONFIRMATION_REQUIRED, got %v", err)
	}
	if called {
		t.Fatal("refused call must not reach the network")
	}
	if _, err := client.Chatbots.Delete(context.Background(), "123", true); err != nil {
		t.Fatalf("confirmed delete: %v", err)
	}
	if !called {
		t.Fatal("confirmed call must reach the network")
	}
}

func TestPathParamsAreEscaped(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.EscapedPath(); got != "/api/v2/chatbots/a%2Fb" {
			t.Errorf("escaped path = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	if _, err := client.Chatbots.Get(context.Background(), "a/b"); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestEmptyTokenSendsNoAuthHeader(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if _, present := r.Header["Authorization"]; present {
			t.Errorf("Authorization header must be absent, got %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	client.token = ""
	if _, err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestRedirectAcrossOriginsDropsAuth(t *testing.T) {
	var sawAuth string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Different port = different origin, same 127.0.0.1 host base —
		// exactly the case Go's default redirect policy keeps auth for.
		http.Redirect(w, r, target.URL+"/api/v2/me", http.StatusFound)
	}))
	defer source.Close()
	client := NewClient("secret-token", WithBaseURL(source.URL))
	if _, err := client.Me(context.Background()); err != nil {
		t.Fatalf("Me: %v", err)
	}
	if sawAuth != "" {
		t.Errorf("token leaked across origins: %q", sawAuth)
	}
}

func TestDotAndEmptyPathSegmentsFailLocally(t *testing.T) {
	called := false
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	for _, badID := range []string{"..", ".", ""} {
		_, err := client.Chatbots.Delete(context.Background(), badID, true)
		apiErr, ok := err.(*Error)
		if !ok || apiErr.Code != "INVALID_PATH_PARAM" {
			t.Fatalf("id %q: expected INVALID_PATH_PARAM, got %v", badID, err)
		}
	}
	if called {
		t.Fatal("rejected paths must never reach the network")
	}
}
