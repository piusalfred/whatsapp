package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	StatusResponse struct {
		Success bool `json:"success,omitempty"`
	}

	MessageStatusUpdateRequest struct {
		MessagingProduct string `json:"messaging_product,omitempty"` // always whatsapp
		Status           string `json:"status,omitempty"`            // always read
		MessageID        string `json:"message_id,omitempty"`
	}
)

// MarkMessageRead sends a read receipt for a message.
// When you receive an incoming message from Webhooks, you can use the /messages endpoint
// to mark the message as read by changing its status to read. Messages marked as read
// display two blue check marks alongside their timestamp:
// We recommend marking incoming messages as read within 30 days of receipt. You cannot mark
// outgoing messages you sent as read. Marking a message as read will also mark earlier
// messages in the conversation as read.
func MarkMessageRead(ctx context.Context, client *http.Client, url, token string, message *Message) (*StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	q := req.URL.Query()
	q.Add("access_token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
