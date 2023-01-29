package whatsapp

type (

	// Response is the response from the WhatsApp server
	//
	//	{
	//	  "messaging_product": "whatsapp",
	//	  "contacts": [{
	//	      "input": "PHONE_NUMBER",
	//	      "wa_id": "WHATSAPP_ID",
	//	    }]
	//	  "messages": [{
	//	      "id": "wamid.ID",
	//	    }]
	//	}
	MessageID struct {
		ID string `json:"id"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappId string `json:"wa_id"`
	}

	Response struct {
		MessagingProduct string            `json:"messaging_product"`
		Contacts         []ResponseContact `json:"contacts"`
		Messages         []MessageID       `json:"messages"`
	}

	// Image ...
	// The Cloud API supports media HTTP caching. If you are using a link (link) to a media
	// asset on your server (as opposed to the ID (id) of an asset you have uploaded to our servers),
	// you can instruct us to cache your asset for reuse with future messages by including
	// the headers below in your server response when we request the asset. If none of these
	// headers are included, we will not cache your asset.
	// Cache-Control: <CACHE_CONTROL>
	// Last-Modified: <LAST_MODIFIED>
	// ETag: <ETAG>
	// Cache-Control
	// The Cache-Control header tells us how to handle asset caching. We support the following directives:
	//
	// max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
	// messages until this time is exceeded, after which we will request the asset again, if needed.
	// Example: Cache-Control: max-age=604800.
	//
	// no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
	// is different from a previous response. Requires the Last-Modified header.
	// Example: Cache-Control: no-cache.
	//
	// no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.
	//
	// private: Indicates that the asset is personalized for the recipient and should not be cached.
	//
	// Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache.
	// If the Last-Modified value is different from a previous response and Cache-Control: no-cache is included
	// in the response, we will update our cached version of the asset with the asset in the response.
	// Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

	// ETag
	// The ETag header is a unique string that identifies a specific version of an asset.
	// Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified
	// headers are not included in the response. In this case, we will cache the asset according to our own,
	//internal logic (which we do not disclose).
	Image struct {
		Link string `json:"link,omitempty"`
		ID   string `json:"id,omitempty"`
	}

	// InteractiveMessage ...
	InteractiveMessage struct {
		Type   string `json:"type"`
		Header struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"header"`
		Body struct {
			Text string `json:"text"`
		} `json:"body"`
		Footer struct {
			Text string `json:"text"`
		} `json:"footer"`
		Action struct {
			Button   string `json:"button"`
			Sections []struct {
				Title string `json:"title"`
				Rows  []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"rows"`
			} `json:"sections"`
		} `json:"action"`
	}
)
