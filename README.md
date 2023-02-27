# whatsapp


## webhooks example

```go

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/models"
	"github.com/piusalfred/whatsapp/webhooks"

	// julieschmidt/httprouter
	"github.com/julienschmidt/httprouter"
)

func main() {
	var options []webhooks.ListenerOption

	// Set your very own notification error handler
	options = append(options, webhooks.WithNotificationErrorHandler(
		func(ctx context.Context, request *http.Request, err error) *webhooks.Response {
			fmt.Printf("error received in notification: %v\n", err) //nolint:forbidigo

			return &webhooks.Response{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Body: []byte(err.Error()),
				Skip: false,
			}
		}))

	// Set your very own subscription verifier
	options = append(options, webhooks.WithSubscriptionVerifier(
		func(ctx context.Context, request *webhooks.VerificationRequest) error {
			fmt.Printf("subscription verification request: %+v\n", request) //nolint:forbidigo
			if request.Mode == "subscribe" && request.Challenge == "challenge" {
				return nil
			}
			return fmt.Errorf("invalid subscription verification request\n")
		}))

	// Set other Listener options here as you wish ........

	// init the listener
	listener := webhooks.NewEventListener(options...)

	// What to do when we receive a reaction?
	// all the logic goes here
	listener.OnMessageReaction(func(ctx context.Context, nctx *webhooks.NotificationContext,
		mctx *webhooks.MessageContext, reaction *models.Reaction,
	) error {
		// do something with the reaction
		fmt.Printf("reaction: %+v\n", reaction) //nolint:forbidigo

		return nil
	})

	// You want to handle media? Things like document, audio, video, image and sticker
	// Well add all your logic here.
	listener.OnMediaMessage(func(ctx context.Context, nctx *webhooks.NotificationContext,
		mctx *webhooks.MessageContext, media *models.MediaInfo,
	) error {
		// do something with the media
		fmt.Printf("media info: %+v\n", media) //nolint:forbidigo

		return nil
	})

	// What to do when we receive a notification of any type be it a media or a text you can handle it here
	// in a generic way where all are caught in one place
	// remember the notification error handler you set above? it will be called here to investigate
	// the error you return
	listener.GenericNotificationHandler(
		func(ctx context.Context, writer http.ResponseWriter,
			notification *webhooks.Notification,
		) error {
			// do something with the notification
			fmt.Printf("notification: %+v\n", notification) //nolint:forbidigo

			return nil
		})

	//// create a http server that listen to port 8080 and all notification on the path POST /webhooks
	//// will be handled by a handler created by the listener on path POST /generic_webhooks the generic
	//// handler will be used. on GET /webhooks the subscription verification will be handled

	// I have used httprouter here, but you can use any other router you wish
	router := httprouter.New()
	router.Handler(http.MethodPost, "/webhooks", listener.Handle())
	router.Handler(http.MethodPost, "/generic_webhooks", listener.GenericHandler())
	router.Handler(http.MethodGet, "/webhooks", listener.Verify())

	// start the server
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

```