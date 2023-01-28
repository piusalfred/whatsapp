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

	ResponseContacts struct {
		Contacts []ResponseContact `json:"contacts"`
	}

	Response struct {
		MessagingProduct string            `json:"messaging_product"`
		Contacts         []ResponseContact `json:"contacts"`
		Messages         []MessageID       `json:"messages"`
	}
)
