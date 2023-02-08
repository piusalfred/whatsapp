package whatsapp

// https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages

const (
	BaseURL                = "https://graph.facebook.com/"
	TextMessageType        = "text"
	ReactionMessageType    = "reaction"
	MediaMessageType       = "media"
	LocationMessageType    = "location"
	ContactMessageType     = "contact"
	InteractiveMessageType = "interactive"
)

type (

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, with the exception of reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	// Response is the response from the WhatsApp server
	// Example:
	//		{
	//	  		"messaging_product": "whatsapp",
	//	  		"contacts": [{
	//	      		"input": "PHONE_NUMBER",
	//	      		"wa_id": "WHATSAPP_ID",
	//	    	}]
	//	  		"messages": [{
	//	      		"id": "wamid.ID",
	//	    	}]
	//		}

	// RequestParams are parameters for a request containing headers, query params,
	// Bearer token, Method and the body.
	// These parameters are used to create a *http.Request

	/*
		CacheOptions contains the options on how to send a media message. You can specify either the
		ID or the link of the media. Also it allows you to specify caching options.

		The Cloud API supports media HTTP caching. If you are using a link (link) to a media asset on your
		server (as opposed to the ID (id) of an asset you have uploaded to our servers),you can instruct us
		to cache your asset for reuse with future messages by including the headers below
		in your server response when we request the asset. If none of these headers are included, we will
		not cache your asset.

			Cache-Control: <CACHE_CONTROL>
			Last-Modified: <LAST_MODIFIED>
			ETag: <ETAG>

		CacheControl

		The Cache-Control header tells us how to handle asset caching. We support the following directives:

			max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
			messages until this time is exceeded, after which we will request the asset again, if needed.
			Example: Cache-Control: max-age=604800.

			no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
			is different from a previous response.Requires the Last-Modified header.
			Example: Cache-Control: no-cache.

			no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.

			private: Indicates that the asset is personalized for the recipient and should not be cached.

		LastModified

		Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache. If the Last-Modified value
		is different from a previous response and Cache-Control: no-cache is included in the response,
		we will update our cached version of the asset with the asset in the response.
		Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

		ETag

		The ETag header is a unique string that identifies a specific version of an asset.
		Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified headers
		are not included in the response. In this case, we will cache the asset according to our own, internal
		logic (which we do not disclose).
	*/

	//MediaType is the type of media to send it can be audio, document, image, sticker, or video.

)

//var InternalSendMediaError = errors.New("internal error while sending media")
