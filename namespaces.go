package sitegpt

import (
	"context"
	"net/http"
	"net/url"
)

// confirmQuery is the query the API requires on destructive calls.
// orEmpty maps a nil input to an empty object. The escalate,
// take-over, and switch-to-ai routes require a JSON body even when
// every field is optional; a nil JSON must mean "{}", not "no body".
func orEmpty(input JSON) JSON {
	if input == nil {
		return JSON{}
	}
	return input
}

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
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots", nil, orEmpty(input))
}

// Update patches a chatbot.
func (a *ChatbotsAPI) Update(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID), nil, orEmpty(input))
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

// DocumentStats summarizes document counts and states. It accepts
// the same filters as ListDocuments (query, source, status, type).
func (a *KnowledgeAPI) DocumentStats(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/stats", query, nil)
}

// AddLinks trains the chatbot on specific URLs.
func (a *KnowledgeAPI) AddLinks(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/links", nil, orEmpty(input))
}

// AddWebsite scrapes and trains on a website.
func (a *KnowledgeAPI) AddWebsite(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/website", nil, orEmpty(input))
}

// AddSitemap trains on every page of a sitemap.
func (a *KnowledgeAPI) AddSitemap(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sitemap", nil, orEmpty(input))
}

// AddYouTube trains on YouTube videos, playlists, or channels.
func (a *KnowledgeAPI) AddYouTube(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/youtube", nil, orEmpty(input))
}

// SetText trains on raw text content.
func (a *KnowledgeAPI) SetText(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/text", nil, orEmpty(input))
}

// ListSources lists connected knowledge sources (Notion, Drive, ...).
func (a *KnowledgeAPI) ListSources(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources", query, nil)
}

// GetDocument returns one trained document. Pass includeContent=true
// in query for the full text.
func (a *KnowledgeAPI) GetDocument(ctx context.Context, chatbotID, documentID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/"+pathParam(documentID), query, nil)
}

// UpdateDocument patches a trained document.
func (a *KnowledgeAPI) UpdateDocument(ctx context.Context, chatbotID, documentID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/"+pathParam(documentID), nil, orEmpty(input))
}

// DeleteDocument removes one trained document. Destructive: confirm
// must be true.
func (a *KnowledgeAPI) DeleteDocument(ctx context.Context, chatbotID, documentID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "delete a knowledge document"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/"+pathParam(documentID), confirmQuery(), nil)
}

// DeleteDocuments bulk-deletes trained documents selected by input.
// Destructive: confirm must be true.
func (a *KnowledgeAPI) DeleteDocuments(ctx context.Context, chatbotID string, input JSON, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "bulk-delete knowledge documents"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/delete", confirmQuery(), orEmpty(input))
}

// ResyncDocuments re-crawls and re-ingests trained documents.
func (a *KnowledgeAPI) ResyncDocuments(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/documents/resync", nil, orEmpty(input))
}

// GetSource returns one connected knowledge source.
func (a *KnowledgeAPI) GetSource(ctx context.Context, chatbotID, connectionID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources/"+pathParam(connectionID), nil, nil)
}

// CreateSource connects a new knowledge source.
func (a *KnowledgeAPI) CreateSource(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources", nil, orEmpty(input))
}

// UpdateSource patches a connected knowledge source.
func (a *KnowledgeAPI) UpdateSource(ctx context.Context, chatbotID, connectionID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources/"+pathParam(connectionID), nil, orEmpty(input))
}

// RevokeSource disconnects a knowledge source. Destructive: confirm
// must be true.
func (a *KnowledgeAPI) RevokeSource(ctx context.Context, chatbotID, connectionID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "revoke a knowledge source"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources/"+pathParam(connectionID), confirmQuery(), nil)
}

// IngestSource triggers ingestion for a connected knowledge source.
func (a *KnowledgeAPI) IngestSource(ctx context.Context, chatbotID, connectionID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/knowledge/sources/"+pathParam(connectionID)+"/ingest", nil, orEmpty(input))
}

// ListSyncJobs lists knowledge ingest and sync jobs.
func (a *KnowledgeAPI) ListSyncJobs(ctx context.Context, chatbotID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/ingest-jobs", query, nil)
}

// GetSyncJob returns one knowledge ingest or sync job.
func (a *KnowledgeAPI) GetSyncJob(ctx context.Context, chatbotID, jobID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/ingest-jobs/"+pathParam(jobID), nil, nil)
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

// Create makes a conversation thread.
func (a *ConversationsAPI) Create(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations", nil, orEmpty(input))
}

// Update patches a conversation (status, tags, assignment, ...).
func (a *ConversationsAPI) Update(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID), nil, orEmpty(input))
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
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/escalate", nil, orEmpty(input))
}

// TakeOver takes over a conversation as a human agent: it sets the
// take-over lock, stops a streaming answer, and posts the
// visitor-visible join notice. input may carry an optional "message".
func (a *ConversationsAPI) TakeOver(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/take-over", nil, orEmpty(input))
}

// SwitchToAI hands an escalated conversation back to the AI. input
// may carry an optional "message".
func (a *ConversationsAPI) SwitchToAI(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/switch-to-ai", nil, orEmpty(input))
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

// Update patches a lead.
func (a *LeadsAPI) Update(ctx context.Context, chatbotID, leadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads/"+pathParam(leadID), nil, orEmpty(input))
}

// Delete removes a lead. Destructive: confirm must be true.
func (a *LeadsAPI) Delete(ctx context.Context, chatbotID, leadID string, confirm bool) (JSON, error) {
	if err := requireConfirmation(confirm, "delete a lead"); err != nil {
		return nil, err
	}
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads/"+pathParam(leadID), confirmQuery(), nil)
}

// RunAction runs a bulk lead action.
func (a *LeadsAPI) RunAction(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/leads/actions", nil, orEmpty(input))
}

// MessagesAPI sends and reads chat messages.
type MessagesAPI struct{ c *Client }

// Send starts a new conversation with a message and returns the reply.
func (a *MessagesAPI) Send(ctx context.Context, chatbotID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/messages", nil, orEmpty(input))
}

// SendToConversation continues an existing conversation.
func (a *MessagesAPI) SendToConversation(ctx context.Context, chatbotID, threadID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/messages", nil, orEmpty(input))
}

// List returns a conversation's messages.
func (a *MessagesAPI) List(ctx context.Context, chatbotID, threadID string, query url.Values) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/messages", query, nil)
}

// Update patches a message (for example, correct an answer).
func (a *MessagesAPI) Update(ctx context.Context, chatbotID, threadID, messageID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPatch, "/api/v2/chatbots/"+pathParam(chatbotID)+"/conversations/"+pathParam(threadID)+"/messages/"+pathParam(messageID), nil, orEmpty(input))
}

// OnboardingAPI is the anonymous agent onboarding flow
// (https://sitegpt.ai/auth.md): create a temporary chatbot before any
// human signs in.
type OnboardingAPI struct{ c *Client }

// Start creates a temporary chatbot workspace from a website URL.
func (a *OnboardingAPI) Start(ctx context.Context, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/onboarding/agent/start", nil, orEmpty(input))
}

// GetWorkspace polls a temporary workspace's state.
func (a *OnboardingAPI) GetWorkspace(ctx context.Context, workspaceID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodGet, "/api/v2/onboarding/workspaces/"+pathParam(workspaceID), nil, nil)
}

// ClaimWorkspace claims an onboarding workspace for a real account.
func (a *OnboardingAPI) ClaimWorkspace(ctx context.Context, workspaceID string, input JSON) (JSON, error) {
	return a.c.Request(ctx, http.MethodPost, "/api/v2/onboarding/workspaces/"+pathParam(workspaceID)+"/claim", nil, orEmpty(input))
}

// DeleteWorkspace deletes an unclaimed onboarding workspace. The
// other SDKs need no confirmation here either: the workspace is a
// temporary artifact of the onboarding flow, not customer data.
func (a *OnboardingAPI) DeleteWorkspace(ctx context.Context, workspaceID string) (JSON, error) {
	return a.c.Request(ctx, http.MethodDelete, "/api/v2/onboarding/workspaces/"+pathParam(workspaceID), nil, nil)
}
