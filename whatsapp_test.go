package whatsapp

import (
	"encoding/json"
	"testing"
)

func TestResponseMarshalJson(t *testing.T) {
	jsonstr := `{"messaging_product": "whatsapp","contacts": [{"input": "PHONE_NUMBER","wa_id": "WHATSAPP_ID"}],"messages": [{"id": "wamid.ID"}]}`
	contacts := []ResponseContact{
		{
			Input:      "123456789",
			WhatsappId: "123456789RTQRWDAR",
		},
	}

	messages := []MessageID{
		{
			ID: "wa:123456789RTQRWDAR:123456789",
		},
	}
	response := &Response{
		MessagingProduct: "whatsapp",
		Contacts:         contacts,
		Messages:         messages,
	}

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(j))

	resulr := &Response{}
	err = json.Unmarshal([]byte(jsonstr), resulr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resulr)
}
