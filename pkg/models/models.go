package models

type (
	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	Text struct {
		PreviewUrl bool   `json:"preview_url,omitempty"`
		Body       string `json:"body,omitempty"`
	}

	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
	}

	Address struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Type        string `json:"type"`
	}

	Addresses struct {
		Addresses []Address `json:"addresses"`
	}

	Email struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	Emails struct {
		Emails []Email `json:"emails"`
	}

	Name struct {
		FormattedName string `json:"formatted_name"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
		Prefix        string `json:"prefix"`
	}

	Org struct {
		Company    string `json:"company"`
		Department string `json:"department"`
		Title      string `json:"title"`
	}

	Phone struct {
		Phone string `json:"phone"`
		Type  string `json:"type"`
		WaID  string `json:"wa_id,omitempty"`
	}

	Phones struct {
		Phones []Phone `json:"phones"`
	}

	Url struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls struct {
		Urls []Url `json:"urls"`
	}

	Contact struct {
		Addresses Addresses `json:"addresses,omitempty"`
		Birthday  string    `json:"birthday"`
		Emails    Emails    `json:"emails,omitempty"`
		Name      Name      `json:"name"`
		Org       Org       `json:"org"`
		Phones    Phones    `json:"phones,omitempty"`
		Urls      Urls      `json:"urls,omitempty"`
	}

	Contacts struct {
		Contacts []Contact `json:"contacts"`
	}

	// Context used to store the context of the conversation.
	// You can send any message as a reply to a previous message in a conversation by including
	// the previous message's ID in the context object.
	// The recipient will receive the new message along with a contextual bubble that displays
	// the previous message's content.
	// Recipients will not see a contextual bubble if:
	//    - replying with a template message ("type":"template")
	//    - replying with an image, video, PTT, or audio, and the recipient is on KaiOS
	// These are known bugs which we are addressing.
	Context struct {
		MessageID string `json:"message_id"`
	}
	MediaInfo struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
	}

	/*
	   Media represents a media object. This object is used to send media messages to WhatsApp users. It contains the following fields:

	     - ID, id (string). Required when type is audio, document, image, sticker, or video and you are not using a link.
	       The media object ID. Do not use this field when message type is set to text.

	     - Link, link (string). Required when type is audio, document, image, sticker, or video and you are not using an uploaded
	       media ID (i.e. you are hosting the media asset on your server). The protocol and URL of the media to be sent. Use only
	       with HTTP/HTTPS URLs. Do not use this field when message type is set to text.

	     - Cloud API users only:

	     - See Media HTTP Caching if you would like us to cache the media asset for future messages.

	     - When we request the media asset from your server you must indicate the media's MIME type by including the
	       Content-Type HTTP header. For example: Content-Type: video/mp4. See Supported Media Types for a list of supported
	       media and their MIME types.

	     - Caption, caption (string). For On-Premises API users on v2.41.2 or newer, this field is required when type is audio,
	       document, image, or video and is limited to 1024 characters. Optional. Describes the specified image, document, or
	       video media. Do not use with audio or sticker media.

	     - Filename, filename (string). Optional. Describes the filename for the specific document. Use only with document media.
	       The extension of the filename will specify what format the document is displayed as in WhatsApp.

	     - Provider, provider (string). Optional. Only used for On-Premises API. This path is optionally used with a link when the
	       HTTP/HTTPS link is not directly accessible and requires additional configurations like a bearer token. For information
	       on configuring providers, see the Media Providers documentation.
	*/
	Media struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
		Provider string `json:"provider,omitempty"`
	}
)
