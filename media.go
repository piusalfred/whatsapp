package whatsapp

import "time"

//Here’s a list of the currently supported media types. Check out Supported Media Types for more information.
// Audio (<16 MB) – ACC, MP4, MPEG, AMR, and OGG formats
// Documents (<100 MB) – text, PDF, Office, and Open Office formats
// Images (<5 MB) – JPEG and PNG formats
// Video (<16 MB) – MP4 and 3GP formats
// Stickers (<100 KB) – WebP format

const (
	MaxAudioSize         = 16 * 1024 * 1024
	MaxDocSize           = 100 * 1024 * 1024
	MaxImageSize         = 5 * 1024 * 1024
	MaxVideoSize         = 16 * 1024 * 1024
	MaxStickerSize       = 100 * 1024
	UploadedMediaTTL     = 30 * 24 * time.Hour
	MediaDownloadLinkTTL = 5 * time.Minute
)

type MediaInfo struct {
	URL              string `json:"url,omitempty"`
	MimeType         string `json:"mime_type,omitempty"`
	Sha256           string `json:"sha256,omitempty"`
	FileSize         int    `json:"file_size,omitempty"`
	Id               string `json:"id,omitempty"`
	MessagingProduct string `json:"messaging_product,omitempty"`
}

//Be sure to keep the following in mind:

// Uploaded media only lasts thirty days
// Generated download URLs only last five minutes
// Always save the media ID when you upload a file
// Here’s a list of the currently supported media types. Check out Supported Media Types for more information.

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
type Media struct {
	ID       string `json:"id,omitempty"`
	Link     string `json:"link,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	Provider string `json:"provider,omitempty"`
}
