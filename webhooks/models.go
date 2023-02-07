package webhooks

type (
	// Media represents a media object
	// //"caption": "CAPTION",
	// "mime_type": "image/jpeg",
	// "sha256": "IMAGE_HASH",
	// "id": "ID"
	Media struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
	}
)
