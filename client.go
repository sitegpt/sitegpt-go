// Package sitegpt is the official Go client for the SiteGPT API v2.
//
// It is a thin, zero-dependency mirror of the Python SDK
// (pypi.org/project/sitegpt): namespaced helpers over a generic
// Request method, returning decoded JSON as map[string]any. Everything
// not covered by a helper is reachable through Request — the OpenAPI
// document at https://sitegpt.ai/api/v2/openapi.json is the contract.
package sitegpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production SiteGPT deployment.
const DefaultBaseURL = "https://sitegpt.ai"

// SDKVersion is reported in the User-Agent header.
const SDKVersion = "0.1.0"

// Error is a structured SiteGPT API error.
type Error struct {
	Status  int
	Code    string
	Message string
	// Hint is a human-readable next step. The server implementation
	// sends it even though the OpenAPI schema omits it.
	Hint string
	// Details is the schema-documented structured error context —
	// arbitrary JSON per the contract (object, array, or scalar);
	// nil when absent.
	Details any
}

func (e *Error) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("sitegpt: %s (%s, HTTP %d): %s", e.Message, e.Code, e.Status, e.Hint)
	}
	return fmt.Sprintf("sitegpt: %s (%s, HTTP %d)", e.Message, e.Code, e.Status)
}

// JSON is a decoded API response body.
type JSON = map[string]any

// Client talks to the SiteGPT API v2. Construct with NewClient.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	Chatbots      *ChatbotsAPI
	Knowledge     *KnowledgeAPI
	Conversations *ConversationsAPI
	Leads         *LeadsAPI
	Messages      *MessagesAPI
	Onboarding    *OnboardingAPI
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL points the client at a non-production deployment.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient replaces the default *http.Client (10s timeout).
// The cross-origin auth-stripping guarantee survives the swap (Codex
// round 7: most callers only want a timeout and must not silently
// lose it): the client is shallow-copied and its redirect hook is
// wrapped — stripping runs first, then the caller's own CheckRedirect
// (or Go's default 10-redirect limit) decides whether to follow.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		wrapped := *httpClient
		callerCheck := httpClient.CheckRedirect
		wrapped.CheckRedirect = func(request *http.Request, via []*http.Request) error {
			origin := via[0].URL
			if request.URL.Scheme != origin.Scheme || request.URL.Host != origin.Host {
				request.Header.Del("Authorization")
			}
			if callerCheck != nil {
				return callerCheck(request, via)
			}
			if len(via) >= 10 {
				return fmt.Errorf("sitegpt: stopped after 10 redirects")
			}
			return nil
		}
		c.httpClient = &wrapped
	}
}

// NewClient builds a client from a SiteGPT API token (create one on
// the dashboard's Agents page, or through the anonymous onboarding
// flow documented at https://sitegpt.ai/auth.md).
//
// Cross-origin redirects never leak the token: Go's http.Client strips
// the Authorization header when a redirect changes host (the same
// guarantee the Python SDK implements by hand).
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			// Go's default only strips Authorization when the redirect
			// leaves the domain (subdomains and port changes keep it) —
			// Codex round 1: enforce the documented guarantee exactly:
			// any scheme/host/port change drops the header.
			CheckRedirect: stripAuthOnOriginChange,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Chatbots = &ChatbotsAPI{c}
	c.Knowledge = &KnowledgeAPI{c}
	c.Conversations = &ConversationsAPI{c}
	c.Leads = &LeadsAPI{c}
	c.Messages = &MessagesAPI{c}
	c.Onboarding = &OnboardingAPI{c}
	return c
}

// Request performs one API call. path must start with "/"; query may
// be nil; body may be nil and is JSON-encoded otherwise. Non-2xx
// responses return *Error.
func (c *Client) Request(ctx context.Context, method, path string, query url.Values, body any) (JSON, error) {
	// Absolute path only (Codex round 3): "api/v2/me" without the
	// leading slash would concatenate into "https://sitegpt.aiapi/..."
	// — authenticated traffic to a stranger's host on a typo.
	if !strings.HasPrefix(path, "/") {
		return nil, &Error{
			Status:  0,
			Code:    "INVALID_PATH",
			Message: "path must start with '/'",
			Hint:    "Pass API paths like \"/api/v2/me\".",
		}
	}
	// One choke point for path hygiene (Codex round 2): url.PathEscape
	// leaves "", "." and ".." unescaped, so a hostile or buggy ID could
	// collapse the URL toward a parent route — the risky case being a
	// confirmed nested delete with an ID of "..". Reject locally.
	for _, segment := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." {
			return nil, &Error{
				Status:  0,
				Code:    "INVALID_PATH_PARAM",
				Message: "path contains an empty or dot segment",
				Hint:    "IDs must be non-empty and must not be '.' or '..'.",
			}
		}
	}
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("sitegpt: encoding request body: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, fmt.Errorf("sitegpt: building request: %w", err)
	}
	// No header at all when the token is empty (Codex round 1): the
	// account-less flows (Onboarding.Start, Health) are anonymous, and
	// "Bearer " with no credential is malformed — auth middleware can
	// reject it before the public handler runs.
	if c.token != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "sitegpt-go/"+SDKVersion)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("sitegpt: %w", err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("sitegpt: reading response: %w", err)
	}
	// The documented contract is non-2xx => *Error, and the
	// implementation must match it exactly (Codex round 6): with a
	// WithHTTPClient that does not follow redirects
	// (http.ErrUseLastResponse), a 3xx would otherwise return JSON, nil
	// for an uncompleted call.
	success := response.StatusCode >= 200 && response.StatusCode <= 299
	var decoded JSON
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &decoded); err != nil {
			if !success {
				return nil, &Error{Status: response.StatusCode, Code: "HTTP_ERROR", Message: http.StatusText(response.StatusCode)}
			}
			return nil, fmt.Errorf("sitegpt: response is not JSON (HTTP %d)", response.StatusCode)
		}
	}
	if !success {
		apiError := &Error{Status: response.StatusCode, Code: "HTTP_ERROR", Message: http.StatusText(response.StatusCode)}
		if errorNode, ok := decoded["error"].(map[string]any); ok {
			if code, ok := errorNode["code"].(string); ok {
				apiError.Code = code
			}
			if message, ok := errorNode["message"].(string); ok {
				apiError.Message = message
			}
			if hint, ok := errorNode["hint"].(string); ok {
				apiError.Hint = hint
			}
			if details, present := errorNode["details"]; present {
				apiError.Details = details
			}
		}
		return nil, apiError
	}
	return decoded, nil
}

// stripAuthOnOriginChange removes the Authorization header whenever a
// redirect changes scheme, host, or port. The same stripping wraps any
// client passed through WithHTTPClient, so the guarantee holds on
// every constructor path.
func stripAuthOnOriginChange(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("sitegpt: stopped after 10 redirects")
	}
	origin := via[0].URL
	if request.URL.Scheme != origin.Scheme || request.URL.Host != origin.Host {
		request.Header.Del("Authorization")
	}
	return nil
}

// requireConfirmation guards destructive helpers: the API requires
// confirm=true, and the SDK keeps that intent explicit instead of
// defaulting it.
func requireConfirmation(confirm bool, action string) error {
	if !confirm {
		return &Error{
			Status:  0,
			Code:    "CONFIRMATION_REQUIRED",
			Message: "Refusing to " + action + " without explicit confirmation",
			Hint:    "Pass confirm=true to " + action + " — the API requires confirm=true for destructive operations.",
		}
	}
	return nil
}

func pathParam(value string) string {
	return url.PathEscape(value)
}

// Me returns the authenticated account.
func (c *Client) Me(ctx context.Context) (JSON, error) {
	return c.Request(ctx, http.MethodGet, "/api/v2/me", nil, nil)
}

// Health checks API availability.
func (c *Client) Health(ctx context.Context) (JSON, error) {
	return c.Request(ctx, http.MethodGet, "/api/v2/health", nil, nil)
}
