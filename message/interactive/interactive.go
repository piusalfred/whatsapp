package interactive

import "github.com/piusalfred/whatsapp/message/media"

const (
	MessageTypeLocationRequest       MessageType = "location_request_message"
	MessageTypeCTAURL                MessageType = "cta_url"
	MessageTypeButton                MessageType = "button"
	MessageTypeFlow                  MessageType = "flow"
	MessageTypeAddressMessage        MessageType = "address_message"
	MessageTypeCarousel              MessageType = "carousel"
	MessageTypeCallPermissionRequest MessageType = "call_permission_request"
	MessageTypeList                  MessageType = "list"
	ActionSendLocation         ActionName = "send_location"
	ActionCTAURL               ActionName = "cta_url"
	ActionButtonReply          ActionName = "reply"
	ActionFlow                 ActionName = "flow"
)

type (
	MessageType string
	ActionName string

	Button struct {
		Type  string       `json:"type,omitempty"`
		Title string       `json:"title,omitempty"`
		ID    string       `json:"id,omitempty"`
		Reply *ReplyButton `json:"reply,omitempty"`
	}

	ReplyButton struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	// SectionRow contains information about a row in an interactive section.
	SectionRow struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	Product struct {
		RetailerID string `json:"product_retailer_id,omitempty"`
	}

	Section struct {
		Title        string        `json:"title,omitempty"`
		ProductItems []*Product    `json:"product_items,omitempty"`
		Rows         []*SectionRow `json:"rows,omitempty"`
	}

	Action struct {
		Button            string            `json:"button,omitempty"`
		Buttons           []*Button         `json:"buttons,omitempty"`
		CatalogID         string            `json:"catalog_id,omitempty"`
		ProductRetailerID string            `json:"product_retailer_id,omitempty"`
		Sections          []*Section        `json:"sections,omitempty"`
		Name              string            `json:"name,omitempty"`
		Parameters        *ActionParameters `json:"parameters,omitempty"`
		Cards             []*Card           `json:"cards,omitempty"`
	}

	Card struct {
		CardIndex int         `json:"card_index,omitempty"`
		Type      string      `json:"type,omitempty"`
		Action    *CardAction `json:"action,omitempty"`
		Header    *CardHeader `json:"header,omitempty"`
		Body      *CardBody   `json:"body,omitempty"`
	}

	CardAction struct {
		Name              string         `json:"name,omitempty"`
		ProductRetailerID string         `json:"product_retailer_id,omitempty"`
		CatalogID         string         `json:"catalog_id,omitempty"`
		Parameters        map[string]any `json:"parameters,omitempty"`
		Buttons           []*CardButton  `json:"buttons,omitempty"`
	}

	CardButton struct {
		Type       string          `json:"type,omitempty"`
		QuickReply *CardQuickReply `json:"quick_reply,omitempty"`
	}

	CardQuickReply struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	CardHeader struct {
		Type  string           `json:"type,omitempty"`
		Image *CardHeaderImage `json:"image,omitempty"`
		Video *CardHeaderVideo `json:"video,omitempty"`
	}

	CardBody struct {
		Text string `json:"text,omitempty"`
	}

	CardHeaderImage struct {
		Link string `json:"link,omitempty"`
	}

	CardHeaderVideo struct {
		Link string `json:"link,omitempty"`
	}

	ActionParameters struct {
		DisplayText        string             `json:"display_text,omitempty"`
		URL                string             `json:"url,omitempty"`
		FlowMessageVersion string             `json:"flow_message_version"`
		FlowToken          string             `json:"flow_token"`
		FlowID             string             `json:"flow_id"`
		FlowCTA            string             `json:"flow_cta"`
		FlowAction         string             `json:"flow_action"`
		FlowActionPayload  *FlowActionPayload `json:"flow_action_payload"`
		Country            string             `json:"country,omitempty"`
		Values             map[string]any     `json:"values,omitempty"`
		ValidationErrors   map[string]any     `json:"validation_errors,omitempty"`
		SavedAddresses     []*SavedAddress    `json:"saved_addresses,omitempty"`
	}

	SavedAddress struct {
		ID    string         `json:"id,omitempty"`
		Value map[string]any `json:"value,omitempty"`
	}

	FlowActionPayload struct {
		Screen string         `json:"screen"`
		Data   map[string]any `json:"data"`
	}

	Header struct {
		Document *media.Document `json:"document,omitempty"`
		Image    *media.Image    `json:"image,omitempty"`
		Video    *media.Video    `json:"video,omitempty"`
		Text     string    `json:"text,omitempty"`
		Type     string    `json:"type,omitempty"`
	}

	// Footer contains information about an interactive footer.
	Footer struct {
		Text string `json:"text,omitempty"`
	}

	// Body contains information about an interactive body.
	Body struct {
		Text string `json:"text,omitempty"`
	}

	Message struct {
		Type   string  `json:"type,omitempty"`
		Action *Action `json:"action,omitempty"`
		Body   *Body   `json:"body,omitempty"`
		Footer *Footer `json:"footer,omitempty"`
		Header *Header `json:"header,omitempty"`
	}

	Option func(*Message)
)
