package whatsapp

import "time"

//Here’s a list of the currently supported media types. Check out Supported Media Types for more information.
// Audio (<16 MB) – ACC, MP4, MPEG, AMR, and OGG formats
// Documents (<100 MB) – text, PDF, Office, and Open Office formats
// Images (<5 MB) – JPEG and PNG formats
// Video (<16 MB) – MP4 and 3GP formats
// Stickers (<100 KB) – WebP format

const (
	MaxAudioSize         = 16 * 1024 * 1024  // 16 MB
	MaxDocSize           = 100 * 1024 * 1024 // 100 MB
	MaxImageSize         = 5 * 1024 * 1024   // 5 MB
	MaxVideoSize         = 16 * 1024 * 1024  // 16 MB
	MaxStickerSize       = 100 * 1024        // 100 KB
	UploadedMediaTTL     = 30 * 24 * time.Hour
	MediaDownloadLinkTTL = 5 * time.Minute
)

//Be sure to keep the following in mind:

// Uploaded media only lasts thirty days
// Generated download URLs only last five minutes
// Always save the media ID when you upload a file
// Here’s a list of the currently supported media types. Check out Supported Media Types for more information.
