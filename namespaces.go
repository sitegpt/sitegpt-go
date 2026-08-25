package sitegpt

import (
	"context"
	"net/http"
	"net/url"
)

// confirmQuery is the query the API requires on destructive calls.
func confirmQuery() url.Values {
	q := url.Values{}
	q.Set("confirm", "true")
	return q
}

// ChatbotsAPI manages chatbots.
type ChatbotsAPI struct{ c *Client }

// List returns the account's chatbots.
func (a *ChatbotsAPI) List(ctx context.Context) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots", nil, nil)
}

// Get returns one chatbot.
func (a *ChatbotsAPI) Get(ctx context.Context, chatbotID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID), nil, nil)
}

// Create makes a chatbot from the given input fields.
func (a *ChatbotsAPI) Create(ctx context.Context, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots", nil, input)
}

// Update patches a chatbot.
func (a *ChatbotsAPI) Update(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID), nil, input)
}

// Delete removes a chatbot. Destructive: confirm must be true.
func (a *ChatbotsAPI) Delete(ctx context.Context, chatbotID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "delete a chatbot"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID), confirmQuery(), nil)
}

// Dashboard returns the chatbot's dashboard summary.
func (a *ChatbotsAPI) Dashboard(ctx context.Context, chatbotID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/dashboard", nil, nil)
}

// Analytics returns the chatbot's daily engagement series with totals
// and a prior-period comparison: widget opens, messages, reactions,
// conversations started, unique visitors, escalations, and leads.
// query may carry "startDay" and "endDay" (YYYY-MM-DD, UTC); the API
// defaults to the trailing 30 days. Accounts without the analytics
// entitlement get a 403 *Error with code ANALYTICS_LOCKED; the
// insight-derived counters appear only when insights is enabled.
func (a *ChatbotsAPI) Analytics(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/analytics", query, nil)
}

// KnowledgeAPI manages a chatbot's training content.
type KnowledgeAPI struct{ c *Client }

// ListDocuments lists trained documents.
func (a *KnowledgeAPI) ListDocuments(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents", query, nil)
}

// DocumentStats summarizes document counts and states.
func (a *KnowledgeAPI) DocumentStats(ctx context.Context, chatbotID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/stats", nil, nil)
}

// AddLinks trains the chatbot on specific URLs.
func (a *KnowledgeAPI) AddLinks(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/links", nil, input)
}

// AddWebsite scrapes and trains on a website.
func (a *KnowledgeAPI) AddWebsite(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/website", nil, input)
}

// AddSitemap trains on every page of a sitemap.
func (a *KnowledgeAPI) AddSitemap(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sitemap", nil, input)
}

// AddYouTube trains on YouTube videos, playlists, or channels.
func (a *KnowledgeAPI) AddYouTube(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/youtube", nil, input)
}

// SetText trains on raw text content.
func (a *KnowledgeAPI) SetText(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/text", nil, input)
}

// ListSources lists connected knowledge sources (Notion, Drive, ...).
func (a *KnowledgeAPI) ListSources(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources", query, nil)
}

// ConversationsAPI reads and manages visitor conversations.
type ConversationsAPI struct{ c *Client }

// List returns conversations, newest first.
func (a *ConversationsAPI) List(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations", query, nil)
}

// Get returns one conversation with its messages.
func (a *ConversationsAPI) Get(ctx context.Context, chatbotID, threadID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID), nil, nil)
}

// Delete removes a conversation. Destructive: confirm must be true.
func (a *ConversationsAPI) Delete(ctx context.Context, chatbotID, threadID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "delete a conversation"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID), confirmQuery(), nil)
}

// Escalate hands the conversation to a human agent.
func (a *ConversationsAPI) Escalate(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/escalate", nil, input)
}

// LeadsAPI reads and manages captured leads.
type LeadsAPI struct{ c *Client }

// List returns captured leads.
func (a *LeadsAPI) List(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads", query, nil)
}

// Get returns one lead.
func (a *LeadsAPI) Get(ctx context.Context, chatbotID, leadID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads/"+pathParam(leadID), nil, nil)
}

// Delete removes a lead. Destructive: confirm must be true.
func (a *LeadsAPI) Delete(ctx context.Context, chatbotID, leadID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "delete a lead"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads/"+pathParam(leadID), confirmQuery(), nil)
}

// MessagesAPI sends and reads chat messages.
type MessagesAPI struct{ c *Client }

// Send starts a new conversation with a message and returns the reply.
func (a *MessagesAPI) Send(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/messages", nil, input)
}

// SendToConversation continues an existing conversation.
func (a *MessagesAPI) SendToConversation(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/messages", nil, input)
}

// OnboardingAPI is the anonymous agent onboarding flow
// (https://sitegpt.ai/auth.md): create a temporary chatbot before any
// human signs in.
type OnboardingAPI struct{ c *Client }

// Start creates a temporary chatbot workspace from a website URL.
func (a *OnboardingAPI) Start(ctx context.Context, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/onboarding/agent/start", nil, input)
}

// GetWorkspace polls a temporary workspace's state.
func (a *OnboardingAPI) GetWorkspace(ctx context.Context, workspaceID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/onboarding/workspaces/"+pathParam(workspaceID), nil, nil)
}
