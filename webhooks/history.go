//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks

import werrors "github.com/piusalfred/whatsapp/pkg/errors"

// HistorySyncContext carries the notification-level metadata for a history
// sync webhook. It is passed to the history handler callback.
type HistorySyncContext struct {
	NotificationObject string // "whatsapp_business_account"
	EntryID            string // WABA ID
	EntryTime          int64  // UNIX timestamp
}

// HistoryEntry represents a single chat history webhook payload. Each entry
// contains metadata about the sync phase and progress, plus a set of message
// threads. Threads are delivered in chunks identified by chunk_order;
// use phase and progress to track overall sync completion (progress == 100
// means all history has been delivered).
type HistoryEntry struct {
	Metadata HistoryMetadata `json:"metadata"`
	Threads  []HistoryThread `json:"threads,omitempty"`
	Errors   []HistoryError  `json:"errors,omitempty"` // non-nil when sharing is declined
}

// HistoryMetadata describes the current sync phase and progress.
// Phase values:
//
//	0 — messages from day 0 (onboarding) through day 1
//	1 — messages from day 1 through day 90
//	2 — messages from day 90 through day 180
//
// Progress ranges 0–100. A value of 100 indicates sync is complete.
// ChunkOrder is a sequential index; chunks may arrive out of order.
type HistoryMetadata struct {
	Phase      int `json:"phase"`
	ChunkOrder int `json:"chunk_order"`
	Progress   int `json:"progress"`
}

// HistoryThread groups messages exchanged with a single WhatsApp user.
// The ID is the user's phone number.
type HistoryThread struct {
	ID       string           `json:"id"`
	Messages []HistoryMessage `json:"messages"`
}

// HistoryMessage represents a single message within a history thread. The
// From field indicates who sent it (the business or the user). Media messages
// arrive in two steps: first as type "media_placeholder" with no content, then
// a separate webhook delivers the actual media details.
type HistoryMessage struct {
	From           string                `json:"from"`
	ID             string                `json:"id"`
	Timestamp      string                `json:"timestamp"`
	Type           string                `json:"type"`
	Text           *HistoryText          `json:"text,omitempty"`
	Image          *HistoryMedia         `json:"image,omitempty"`
	Video          *HistoryMedia         `json:"video,omitempty"`
	Audio          *HistoryMedia         `json:"audio,omitempty"`
	Document       *HistoryDocument      `json:"document,omitempty"`
	Sticker        *HistoryMedia         `json:"sticker,omitempty"`
	Contacts       []HistoryContact      `json:"contacts,omitempty"`
	Location       *HistoryLocation      `json:"location,omitempty"`
	Reaction       *HistoryReaction      `json:"reaction,omitempty"`
	Order          *HistoryOrder         `json:"order,omitempty"`
	Interactive    *HistoryInteractive   `json:"interactive,omitempty"`
	Button         *HistoryButton        `json:"button,omitempty"`
	System         *HistorySystemMessage `json:"system,omitempty"`
	HistoryContext *HistoryContext       `json:"history_context,omitempty"`
}

// HistoryText is the content of a text message in history.
type HistoryText struct {
	Body string `json:"body"`
}

// HistoryMedia describes a media message (image, video, audio, sticker)
// in a history webhook. The ID field is the media asset ID; only present
// for messages sent within 14 days of onboarding.
type HistoryMedia struct {
	ID       string `json:"id,omitempty"`
	Caption  string `json:"caption,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
}

// HistoryDocument describes a document message in history.
type HistoryDocument struct {
	ID       string `json:"id,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
}

// HistoryContact is a contact shared in a history message.
type HistoryContact struct {
	Name *HistoryContactName `json:"name,omitempty"`
	URLs []HistoryContactURL `json:"urls,omitempty"`
	Orgs []HistoryContactOrg `json:"orgs,omitempty"`
}

// HistoryContactName is the structured name of a contact.
type HistoryContactName struct {
	FormattedName string `json:"formatted_name"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
}

// HistoryContactURL is a URL associated with a contact.
type HistoryContactURL struct {
	URL  string `json:"url"`
	Type string `json:"type,omitempty"`
}

// HistoryContactOrg is an organization associated with a contact.
type HistoryContactOrg struct {
	Company string `json:"company,omitempty"`
	Title   string `json:"title,omitempty"`
}

// HistoryLocation is a shared location in history.
type HistoryLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

// HistoryReaction is a reaction to a message in history.
type HistoryReaction struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

// HistoryOrder describes an order shared in history.
type HistoryOrder struct {
	CatalogID    string                    `json:"catalog_id"`
	Text         string                    `json:"text,omitempty"`
	ProductItems []HistoryOrderProductItem `json:"product_items,omitempty"`
}

// HistoryOrderProductItem is a product item within an order.
type HistoryOrderProductItem struct {
	ProductRetailerID string `json:"product_retailer_id"`
	Quantity          int    `json:"quantity"`
	ItemPrice         int    `json:"item_price"`
	Currency          string `json:"currency"`
}

// HistoryInteractive describes an interactive message in history.
type HistoryInteractive struct {
	Type        string              `json:"type"`
	ButtonReply *HistoryButtonReply `json:"button_reply,omitempty"`
	ListReply   *HistoryListReply   `json:"list_reply,omitempty"`
}

// HistoryButtonReply is a button reply within an interactive message.
type HistoryButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// HistoryListReply is a list reply within an interactive message.
type HistoryListReply struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// HistoryButton describes a button message in history.
type HistoryButton struct {
	Payload string `json:"payload"`
	Text    string `json:"text"`
}

// HistorySystemMessage describes a system-generated message in history
// (e.g., customer identity change).
type HistorySystemMessage struct {
	Body     string `json:"body"`
	Identity string `json:"identity,omitempty"`
	NewWaID  string `json:"new_wa_id,omitempty"`
	WaID     string `json:"wa_id,omitempty"`
	Type     string `json:"type"`
	Customer string `json:"customer,omitempty"`
}

// HistoryContext provides the delivery status of a message in history.
// Status values: DELIVERED, ERROR, PENDING, PLAYED, READ, SENT.
type HistoryContext struct {
	Status string `json:"status"`
}

// HistoryError describes a history sync failure (e.g., sharing declined).
// Convert to *werrors.Error via the Error() method.
type HistoryError struct {
	Code   int                `json:"code"`
	Title  string             `json:"title"`
	Detail string             `json:"message"`
	Data   *werrors.ErrorData `json:"error_data,omitempty"`
}
