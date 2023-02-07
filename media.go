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
