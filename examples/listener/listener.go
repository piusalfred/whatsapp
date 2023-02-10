package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/piusalfred/whatsapp/webhooks"
)

var _ webhooks.EventHandler = (*handler)(nil)

type handler struct {
	// it can be db conections, etc
	// connection to notify the user
	// etc etc
}

func (h *handler) HandleError(ctx context.Context, writer http.ResponseWriter, request *http.Request, err error) error {
	if err != nil {
		os.Stdout.WriteString(err.Error())
	}

	os.Stdout.WriteString("error is nil")
	return nil
}

func (h *handler) HandleEvent(ctx context.Context, writer http.ResponseWriter, request *http.Request, notification *webhooks.Notification) error {
	os.Stdout.WriteString("HandleEvent")
	jsonb, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(jsonb)
	return nil
}

//	curl -X POST --location "http://localhost:8080/webhooks" \
//	   -H "Content-Type: application/json" \
//	   -d "{
//	         \"object\": \"whatsapp_business_account\",
//	         \"entry\": [{
//	           \"id\": \"WHATSAPP_BUSINESS_ACCOUNT_ID\",
//	           \"changes\": [{
//	             \"value\": {
//	               \"messaging_product\": \"whatsapp\",
//	               \"metadata\": {
//	                 \"display_phone_number\": \"PHONE_NUMBER\",
//	                 \"phone_number_id\": \"PHONE_NUMBER_ID\"
//	               },
//	               \"contacts\": [{
//	                 \"profile\": {
//	                   \"name\": \"NAME\"
//	                 },
//	                 \"wa_id\": \"PHONE_NUMBER_ID\"
//	               }],
//	               \"messages\": [{
//	                 \"from\": \"PHONE_NUMBER\",
//	                 \"id\": \"wamid.ID\",
//	                 \"timestamp\": \"TIMESTAMP\",
//	                 \"text\": {
//	                   \"body\": \"MESSAGE_BODY\"
//	                 },
//	                 \"type\": \"text\"
//	               }]
//	             },
//	             \"field\": \"messages\"
//	           }]
//	         }]
//	       }"
func main() {
	// Create a new handler
	handler := &handler{}
	ls := webhooks.NewEventListener(handler)
	mux := http.NewServeMux()
	mux.Handle("/webhooks", ls)

	// Create a new server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}
